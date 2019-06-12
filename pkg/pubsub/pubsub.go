package pubsub

import (
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
)

// Publisher is a publisher to broadcast messages to all subscribers
type Publisher struct {
	sync.RWMutex                         // protect m
	m            map[Subcriber]TopicFunc // hold all of subcribers

	timeout   time.Duration // send topic timeout
	bufferLen int           // buffer length of each subcriber channel
}

// Subcriber is exported
type Subcriber chan interface{}

// TopicFunc is a function to check if the subscriber has interest in one topic message
type TopicFunc func(v interface{}) bool

// NewPublisher initialize a publisher
func NewPublisher(timeout time.Duration, bufferLen int) *Publisher {
	return &Publisher{
		m:         make(map[Subcriber]TopicFunc),
		timeout:   timeout,
		bufferLen: bufferLen,
	}
}

// SubcribeAll adds a new subscriber that receive all messages.
func (p *Publisher) SubcribeAll() Subcriber {
	return p.Subcribe(nil)
}

// Subcribe adds a new subscriber that filters messages sent by a topic.
func (p *Publisher) Subcribe(tf TopicFunc) Subcriber {
	ch := make(Subcriber, p.bufferLen)
	p.Lock()
	p.m[ch] = tf
	p.Unlock()
	return ch
}

// Evict removes the specified subscriber from receiving any more messages.
func (p *Publisher) Evict(sub Subcriber) {
	p.Lock()
	delete(p.m, sub)
	close(sub)
	p.Unlock()
}

// Publish broadcast message to all subscribers simultaneously
func (p *Publisher) Publish(v interface{}) {
	p.RLock()
	defer p.RUnlock()

	var wg sync.WaitGroup
	wg.Add(len(p.m))
	// broadcasting with concurrency
	for sub, tf := range p.m {
		go func(sub Subcriber, v interface{}, tf TopicFunc) {
			defer wg.Done()
			p.send(sub, v, tf)
		}(sub, v, tf)
	}
	wg.Wait()
}

// NumHitTopic return nb of waitting subscribers who cares about the specified value.
func (p *Publisher) NumHitTopic(v interface{}) (n int) {
	p.RLock()
	defer p.RUnlock()
	for _, tf := range p.m {
		if tf == nil || tf(v) {
			n++
		}
	}
	return
}

func (p *Publisher) send(sub Subcriber, v interface{}, tf TopicFunc) {
	// if a subcriber setup topic filter func and not matched by the topic filter
	// skip send message to this subcriber
	if tf != nil && !tf(v) {
		return
	}

	// send with timeout
	if p.timeout > 0 {
		select {
		case sub <- v:
		case <-time.After(p.timeout):
			log.Println("send to subcriber timeout after", p.timeout.String())
		}
		return
	}

	// directely send
	sub <- v
}
