package common

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
)

type HighLowSlice struct {
	mu        sync.RWMutex
	d         []*ArticleRef
	MaxSize   int
	high, low int
}

func (s *HighLowSlice) Len() int { return s.high }

func (s *HighLowSlice) High() int { return s.high }

func (s *HighLowSlice) Low() int { return s.low }

func (s *HighLowSlice) String() string {
	return fmt.Sprintf("[%v~%v max:%v data:%v", s.low, s.high, s.MaxSize, s.d)
}

func (s *HighLowSlice) Get(i int) (*ArticleRef, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if i < s.low || i >= s.high {
		return nil, false
	}
	i -= s.low
	return s.d[i], true
}

func (s *HighLowSlice) Set(i int, v *ArticleRef) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if i < s.low {
		return
	}
	i -= s.low
	s.d[i] = v
}

func (s *HighLowSlice) Slice(i, j int, copy bool) (results []*ArticleRef, actualStart, actualEnd int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if j <= s.low {
		return
	}
	if j > s.high {
		j = s.high
	}
	j -= s.low

	if i < s.low {
		i = 0
	} else {
		i -= s.low
	}

	if i > j {
		return
	}

	if copy {
		results = append([]*ArticleRef{}, s.d[i:j]...)
	} else {
		results = s.d[i:j]
	}

	actualStart = i + s.low
	actualEnd = j + s.low
	return
}

func (s *HighLowSlice) Append(v *ArticleRef) ([]*ArticleRef, int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.d = append(s.d, v)
	s.high++

	var purged []*ArticleRef
	if len(s.d) > s.MaxSize {
		p := 1 / float64(len(s.d)-s.MaxSize+1)
		if rand.Float64() > p {
			x := len(s.d) - s.MaxSize
			purged = append([]*ArticleRef{}, s.d[:x]...)

			s.low += x
			copy(s.d, s.d[x:])
			s.d = s.d[:s.MaxSize]
		}
	}
	return purged, s.high
}

func PanicIf(err interface{}, f string, a ...interface{}) {
	if v, ok := err.(bool); ok {
		if v {
			log.Fatalf(f, a...)
		}
		return
	}
	if err != nil {
		f = strings.Replace(f, "%%err", strings.Replace(fmt.Sprint(err), "%", "%%", -1), -1)
		log.Fatalf(f, a...)
	}
}

func ExtractMsgID(msgID string) string {
	if strings.HasPrefix(msgID, "<") && strings.HasSuffix(msgID, ">") {
		msgID = msgID[1 : len(msgID)-1]
		msgID = strings.Split(msgID, "@")[0]
	}
	return msgID
}

func ExtractEmail(from string) string {
	start, end := strings.Index(from, "<"), strings.Index(from, ">")
	if start > -1 && end > -1 && end > start {
		return from[start+1 : end]
	}
	return ""
}

func MsgIDToRawMsgID(msgid string, msgidbuf []byte) [16]byte {
	var x [16]byte
	if msgidbuf != nil {
		copy(x[:], msgidbuf)
	} else {
		copy(x[:], msgid)
	}
	return x
}