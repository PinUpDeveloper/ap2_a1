package provider

import (
	"errors"
	"log"
	"math/rand"
	"time"
)

type SimulatedEmailSender struct {
	FailureRate float64 // 0.0–1.0; default 0.3
	MinLatency  time.Duration
	MaxLatency  time.Duration
}

func NewSimulatedEmailSender() *SimulatedEmailSender {
	return &SimulatedEmailSender{
		FailureRate: 0.3,
		MinLatency:  200 * time.Millisecond,
		MaxLatency:  800 * time.Millisecond,
	}
}

func (s *SimulatedEmailSender) Send(msg EmailMessage) error {
	latency := s.MinLatency + time.Duration(rand.Int63n(int64(s.MaxLatency-s.MinLatency)))
	time.Sleep(latency)

	if rand.Float64() < s.FailureRate {
		log.Printf("[SimulatedProvider] transient failure sending email to %s", msg.To)
		return errors.New("simulated: transient provider failure")
	}

	log.Printf("[SimulatedProvider] ✓ email sent to=%s subject=%q (latency=%s)", msg.To, msg.Subject, latency)
	return nil
}
