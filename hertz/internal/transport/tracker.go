// Package transport 提供 Hertz Server 私有的连接排空能力。
package transport

import (
	"context"
	"io"
	"sync"
	"sync/atomic"

	"github.com/cloudwego/hertz/pkg/network"
)

type connectionState uint32

const (
	connectionIdle connectionState = iota
	connectionServing
	connectionResponsePending
	connectionClosed
)

// Tracker 跟踪由一个 Hertz Server 持有的连接。
type Tracker struct {
	mu       sync.Mutex
	conns    map[*Connection]struct{}
	draining atomic.Bool
}

// NewTracker 创建连接跟踪器。
func NewTracker() *Tracker {
	return &Tracker{conns: make(map[*Connection]struct{})}
}

// Wrap 使用连接跟踪能力装饰 Hertz transport。
func (t *Tracker) Wrap(value network.Transporter) network.Transporter {
	return &trackedTransport{transport: value, tracker: t}
}

// BeginShutdown 进入排空状态并立即关闭空闲连接。
func (t *Tracker) BeginShutdown() {
	if t == nil || !t.draining.CompareAndSwap(false, true) {
		return
	}
	t.mu.Lock()
	idle := make([]*Connection, 0, len(t.conns))
	for conn := range t.conns {
		if connectionState(conn.state.Load()) == connectionIdle {
			idle = append(idle, conn)
		}
	}
	t.mu.Unlock()
	for _, conn := range idle {
		_ = conn.Close()
	}
}

// Draining 报告 Server 是否正在排空连接。
func (t *Tracker) Draining() bool {
	return t != nil && t.draining.Load()
}

func (t *Tracker) track(conn network.Conn) *Connection {
	tracked := &Connection{Conn: conn, owner: t}
	t.mu.Lock()
	t.conns[tracked] = struct{}{}
	t.mu.Unlock()
	return tracked
}

func (t *Tracker) remove(conn *Connection) {
	t.mu.Lock()
	delete(t.conns, conn)
	t.mu.Unlock()
}

// Connection 是携带请求处理状态的 Hertz 连接。
type Connection struct {
	network.Conn
	owner *Tracker
	state atomic.Uint32
	once  sync.Once
}

// BeginRequest 标记连接正在执行业务处理器。
func (c *Connection) BeginRequest() {
	if c != nil {
		c.state.Store(uint32(connectionServing))
	}
}

// FinishRequest 标记连接等待 Hertz 刷新最终响应。
func (c *Connection) FinishRequest() {
	if c != nil {
		c.state.CompareAndSwap(uint32(connectionServing), uint32(connectionResponsePending))
	}
}

// Flush 刷新响应；排空期间在完整响应写出后关闭连接。
func (c *Connection) Flush() error {
	err := c.Conn.Flush()
	if err != nil || !c.responseFlushed() {
		return err
	}
	_ = c.Close()
	return nil
}

// Write 保留直接写出路径的响应完成语义。
func (c *Connection) Write(payload []byte) (int, error) {
	written, err := c.Conn.Write(payload)
	if err == nil && written == len(payload) && c.responseFlushed() {
		_ = c.Close()
	}
	return written, err
}

// ReadFrom 保留底层连接支持的零拷贝读写路径。
func (c *Connection) ReadFrom(reader io.Reader) (int64, error) {
	if source, ok := c.Conn.(io.ReaderFrom); ok {
		return source.ReadFrom(reader)
	}
	return io.Copy(struct{ io.Writer }{c.Conn}, reader)
}

// Close 关闭连接并从跟踪器中移除。
func (c *Connection) Close() error {
	if c == nil || c.Conn == nil {
		return nil
	}
	var err error
	c.once.Do(func() {
		c.state.Store(uint32(connectionClosed))
		c.owner.remove(c)
		err = c.Conn.Close()
	})
	return err
}

func (c *Connection) responseFlushed() bool {
	return c.state.CompareAndSwap(uint32(connectionResponsePending), uint32(connectionIdle)) && c.owner.Draining()
}

type trackedTransport struct {
	transport network.Transporter
	tracker   *Tracker
}

func (t *trackedTransport) ListenAndServe(onData network.OnData) error {
	return t.transport.ListenAndServe(func(ctx context.Context, value any) error {
		conn, ok := value.(network.Conn)
		if !ok {
			return onData(ctx, value)
		}
		tracked := t.tracker.track(conn)
		return onData(context.WithValue(ctx, connectionContextKey{}, tracked), tracked)
	})
}

func (t *trackedTransport) Shutdown(ctx context.Context) error {
	return t.transport.Shutdown(ctx)
}

func (t *trackedTransport) Close() error {
	return t.transport.Close()
}

type connectionContextKey struct{}

// FromContext 返回当前请求所属的受跟踪连接。
func FromContext(ctx context.Context) *Connection {
	conn, _ := ctx.Value(connectionContextKey{}).(*Connection)
	return conn
}
