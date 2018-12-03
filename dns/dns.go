package dns

import (
	"fmt"
	"log"
	"net"

	"github.com/pkg/errors"
	"golang.org/x/net/dns/dnsmessage"
)

// Store interface
type Store interface {
	Get(string) ([]dnsmessage.Resource, bool)
	Set(string, dnsmessage.Resource, *dnsmessage.Resource) bool
}

const (
	packetLen int = 512
)

// Service dns service
type Service struct {
	Book Store
	conn *net.UDPConn
}

// NewService new service
func NewService(book Store) *Service {
	return &Service{
		Book: book,
	}
}

// Packet dns packet
type Packet struct {
	addr    net.UDPAddr
	message dnsmessage.Message
}

// Start start dns server
func (s *Service) Start(addr string) {
	var err error
	laddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Fatalf("%+v %s", errors.Wrap(err, "net resolve udp addr err"), addr)
	}
	s.conn, err = net.ListenUDP("udp", laddr)
	if err != nil {
		log.Fatalf("%+v %s", errors.Wrap(err, "net listen on udp err"), laddr)
	}
	defer s.conn.Close()

	for {
		buf := make([]byte, packetLen)
		_, addr, err := s.conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}

		var m dnsmessage.Message
		err = m.Unpack(buf)
		if err != nil {
			log.Println(err)
			continue
		}

		if len(m.Questions) == 0 {
			continue
		}

		go s.query(Packet{*addr, m})
	}
}

func (s *Service) query(p Packet) {
	q := p.message.Questions[0]
	val, ok := s.Book.Get(qString(q))

	if ok {
		p.message.Answers = append(p.message.Answers, val...)
		go sendPacket(s.conn, p.message, p.addr)
	}
}

func sendPacket(conn *net.UDPConn, message dnsmessage.Message, addr net.UDPAddr) {
	packed, err := message.Pack()
	if err != nil {
		log.Println(err)
		return
	}

	_, err = conn.WriteToUDP(packed, &addr)
	if err != nil {
		log.Println(err)
	}
}

// packet to string
func pString(p Packet) string {
	return fmt.Sprint(p.message.ID)
}

// question to string
func qString(q dnsmessage.Question) string {
	b := make([]byte, q.Name.Length+2)
	for i := 0; i < int(q.Name.Length); i++ {
		b[i] = q.Name.Data[i]
	}
	b[q.Name.Length] = uint8(q.Type >> 8)
	b[q.Name.Length+1] = uint8(q.Type)

	return string(b)
}

// NTString name + type to string
func NTString(rName dnsmessage.Name, rType dnsmessage.Type) string {
	b := make([]byte, rName.Length+2)
	for i := 0; i < int(rName.Length); i++ {
		b[i] = rName.Data[i]
	}
	b[rName.Length] = uint8(rType >> 8)
	b[rName.Length+1] = uint8(rType)

	return string(b)
}
