//go:build !ignore || !darwin
// +build !ignore !darwin

package tool

// 当前文件用于 tcp 数据快速拷贝
import (
	"crypto/sha256"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"io"
	"log"
	"net"
	"syscall"
	"time"

	"github.com/unstoppablego/framework/logs"

	"golang.org/x/crypto/chacha20"
)

func CheckCloseByReadOneByte(c net.Conn, err error, readLen int, debug bool) (bool, []byte, int) {
	defer func() {
		if r := recover(); r != nil {
			if debug {
				log.Println("CheckCloseByReadOneByte panic ", r)
			}
			return
		}
	}()
	var one = make([]byte, 1)
	if tc, ok := c.(*tls.Conn); ok {
		len, err := tc.Read(one) //c 语言中 内核代码会去读去，如果有错误的话回直接返回 ,由于外层使用的是 Copy From 函数可能屏蔽了一部分错误信息
		if len == 0 && err != nil {
			if debug {
				log.Println("CheckCloseByReadOneByte conn to tls conn test", err, len)
			}
			return true, nil, 0
		}
		if len == 0 && err == nil {
			if err := connMsgPeekCheck(c); err != nil {
				if debug {
					log.Println("connMsgPeekCheck conn to tls conn test", err, len)
				}
				return true, nil, 0
			}
		}
		return false, one, len
	} else if tc, ok := c.(*net.TCPConn); ok {
		len, err := tc.Read(one)
		if len == 0 && err != nil {
			if debug {
				log.Println("CheckCloseByReadOneByte conn to tcp conn test", err, len)
			}
			return true, nil, 0
		}
		if len == 0 && err == nil {
			if err := connMsgPeekCheck(c); err != nil {
				if debug {
					log.Println("connMsgPeekCheck conn to tcp conn test", err, len)
				}
				return true, nil, 0
			}
		}
		return false, one, len
	} else {
		if debug {
			log.Println("CheckCloseByReadOneByte unknow net conn tpye")
		}
		return true, nil, 0
	}
}

type LimitSpeed interface {
	IsLimit(name string) bool
}
type NoneLimitSpeed struct {
}

func (NoneLimitSpeed) Islimit(name string) bool {
	return false
}

// net conn copy from src to dst and process close conn
func CopyConn(src net.Conn, dst net.Conn, debug bool, limit func() bool, upTraffic func(int)) (allwritten int) {

	defer src.Close()
	defer dst.Close()
	var sized = 32 * 1024 * 2
	written := 0

	defer func() { upTraffic(allwritten) }()

	var arMin int = 1024 * 32 * 12
	needLimit := limit()
	if needLimit {
		arMin = 1024 * 32 * 1
		sized = 1024
	}
	var recvBuf = make([]byte, sized, sized) // 128/4
	var firstBitTorrentCheck bool
	for {
		now := time.Now().Nanosecond()
		rlen, err := src.Read(recvBuf)

		if rlen > 20 && firstBitTorrentCheck == false {
			firstBitTorrentCheck = true
			//明确了TCP BitTorrent protocol 检测方法
			//难道是明文？
			if recvBuf[0] == 19 && string(recvBuf[1:20]) == "BitTorrent protocol" {
				logs.Error("not support BitTorrent protocol")
				return
			}
			if IsNotSupportEmail(recvBuf[0:rlen]) {
				logs.Error("not support Email Send")
				return
			}
		}

		if rlen > 0 {
			//处理写入
			wlen, err := dst.Write(recvBuf[0:rlen])
			if wlen < 0 || rlen < wlen {
				wlen = 0
				if err == nil {
					err = errors.New("write error")
				}
			}
			written += wlen
			allwritten += wlen
			// logs.Info(allwritten)
			//上报流量输出
			if allwritten/1024/1024 > 512 {
				// logs.Info("超过512m上报")
				upTraffic(allwritten)
				allwritten = 0
			}

			if err != nil {
				if debug {
					log.Println("CopyConn dst.Write error", err, src.RemoteAddr()) //打印下未知错误
				}
				return
			}
			if rlen != wlen {
				err = errors.New("rlen != wlen")
				if debug {
					log.Println("CopyConn dst.Write error", err, src.RemoteAddr()) //打印下未知错误
				}
				return
			}
		}

		if err != nil {
			if err == io.EOF || err == net.ErrClosed || err == syscall.EPIPE {
				return
			} else {
				if debug {
					log.Println("CopyConn unknow error", err, src.RemoteAddr()) //打印下未知错误
				}
				return
			}
		} else {
			if rlen == 0 {

				if cs, ok := src.(*Chacha20Stream); ok {
					if closed, one, len := CheckCloseByReadOneByte(cs.Conn, err, int(rlen), debug); !closed {
						if len > 0 {
							if _, err := dst.Write(one[0:len]); err != nil {
								if debug {
									log.Println("CopyConn write one byte err", err)
								}
								return
							}
						}
					} else {
						return
					}
				} else {
					if closed, one, len := CheckCloseByReadOneByte(src, err, int(rlen), debug); !closed {
						if len > 0 {
							if _, err := dst.Write(one[0:len]); err != nil {
								if debug {
									log.Println("CopyConn write one byte err", err)
								}
								return
							}
						}
					} else {
						return
					}
				}

			}
		}

		//limit speed
		if needLimit {
			// logs.Info(fmt.Sprintf("sleep %d ms", wl))
			dif := time.Now().Nanosecond() - now
			// logs.Info(fmt.Sprintf(" %d %d", written, dif))
			//经过了多少毫秒
			waitTime := written*1000000000/arMin - dif
			// logs.Info(fmt.Sprintf(" %d %d %d %d %d", written, arMin, dif, (written * 1000000000 / arMin), waitTime))
			if waitTime > 0 {
				// logs.Info(fmt.Sprintf("sleep %d nas", waitTime))
				time.Sleep(time.Duration(waitTime) * time.Nanosecond)
				written = 0
				// now = time.Now().Nanosecond() / 1e6
			}
		}
	}
}

func CopyConnNoLimit() bool {
	return false
}

// Note that if another goroutine is not actively read()'ing the stream, this will not return EOF until the stream is read and its buffer drained.
func connMsgPeekCheck(connc net.Conn) error {
	return nil
	// var conn net.Conn
	// if tc, ok := connc.(*tls.Conn); ok {
	// 	conn = tc.NetConn()
	// } else if _, ok := connc.(*net.TCPConn); ok {
	// 	conn = connc
	// } else {
	// 	return errors.New("connMsgPeekCheck unknow net conn tpye")
	// }

	// var sysErr error = nil
	// rc, err := conn.(syscall.Conn).SyscallConn()
	// if err != nil {
	// 	return err
	// }
	// err = rc.Read(func(fd uintptr) bool {
	// 	var buf []byte = []byte{0}
	// 	n, _, err := syscall.Recvfrom(int(fd), buf, syscall.MSG_PEEK|syscall.MSG_DONTWAIT) //MSG_DONTWAIT 调用函数时不阻塞
	// 	switch {
	// 	case n == 0 && err == nil:
	// 		sysErr = io.EOF
	// 	case err == syscall.EAGAIN || err == syscall.EWOULDBLOCK: //syscall.EAGAIN 非阻塞模式下调用了阻塞操作 在VxWorks和Windows上，EAGAIN的名字叫做EWOULDBLOCK
	// 		sysErr = nil
	// 	default:
	// 		sysErr = err
	// 	}
	// 	return true
	// })
	// if err != nil {
	// 	return err
	// }
	// return sysErr
}

type Chacha20Stream struct {
	key     []byte
	Nonce   []byte
	encoder *chacha20.Cipher
	decoder *chacha20.Cipher
	Conn    net.Conn
}

func NewChacha20Stream(key string, Nonce int64, conn net.Conn) (*Chacha20Stream, error) {
	key32 := sha256.Sum256([]byte(key))
	s := &Chacha20Stream{key: key32[:], // should be exactly 32 bytes
		Conn: conn}

	var buf = make([]byte, 8)
	s.Nonce = make([]byte, chacha20.NonceSizeX)
	binary.BigEndian.PutUint64(buf, uint64(Nonce))
	copy(s.Nonce, buf)

	var err error
	s.encoder, err = chacha20.NewUnauthenticatedCipher(s.key, s.Nonce)
	if err != nil {
		logs.Error(err)
	}
	s.decoder, err = chacha20.NewUnauthenticatedCipher(s.key, s.Nonce)
	if err != nil {
		logs.Error(err)
	}
	return s, nil
}

func (s *Chacha20Stream) Read(p []byte) (int, error) {
	n, err := s.Conn.Read(p)
	if err != nil || n == 0 {
		return n, err
	}
	dst := make([]byte, n)
	pn := p[:n]
	s.decoder.XORKeyStream(dst, pn)
	copy(pn, dst)
	return n, nil
}

func (s *Chacha20Stream) Write(p []byte) (int, error) {
	dst := make([]byte, len(p))
	s.encoder.XORKeyStream(dst, p)
	return s.Conn.Write(dst)
}

func (s *Chacha20Stream) Close() error {
	return s.Conn.Close()
}

func (s *Chacha20Stream) LocalAddr() net.Addr {
	return s.Conn.LocalAddr()
}

func (s *Chacha20Stream) RemoteAddr() net.Addr {
	return s.Conn.LocalAddr()
}

// func (s )

func (s *Chacha20Stream) SetDeadline(t time.Time) error {
	return s.Conn.SetDeadline(t)
}

func (s *Chacha20Stream) SetReadDeadline(t time.Time) error {
	return s.Conn.SetReadDeadline(t)
}

func (s *Chacha20Stream) SetWriteDeadline(t time.Time) error {
	return s.Conn.SetWriteDeadline(t)
}
