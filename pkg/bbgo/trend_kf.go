package bbgo

import (
	"github.com/sirupsen/logrus"

	"github.com/c9s/bbgo/pkg/indicator"
	"github.com/c9s/bbgo/pkg/types"
)

type TrendKalmanFilter struct {
	types.IntervalWindow

	// MaxGradient is the maximum gradient allowed for the entry.
	MaxGradient float64 `json:"maxGradient"`
	MinGradient float64 `json:"minGradient"`
	Smooth      uint    `json:"smoothLength"`

	estimator *indicator.KalmanFilter

	last, current float64
}

func (s *TrendKalmanFilter) Bind(session *ExchangeSession, orderExecutor *GeneralOrderExecutor) {
	if s.MaxGradient == 0.0 {
		s.MaxGradient = 1.0
	}

	symbol := orderExecutor.Position().Symbol
	s.estimator = &indicator.KalmanFilter{
		IntervalWindow:         s.IntervalWindow,
		AdditionalSmoothWindow: s.Smooth,
	}

	session.MarketDataStream.OnStart(func() {
		if s.estimator.Length() < 2 {
			return
		}

		s.last = s.estimator.Values[s.estimator.Length()-2]
		s.current = s.estimator.Last()
	})

	session.MarketDataStream.OnKLineClosed(types.KLineWith(symbol, s.Interval, func(kline types.KLine) {
		s.estimator.PushK(kline)
		s.last = s.current
		s.current = s.estimator.Last()
	}))
}

func (s *TrendKalmanFilter) Gradient() float64 {
	if s.last > 0.0 && s.current > 0.0 {
		return s.current / s.last
	}
	return 0.0
}

func (s *TrendKalmanFilter) GradientAllowed() bool {
	gradient := s.Gradient()

	logrus.Infof("TrendKalmanFilter %+v current=%f last=%f gradient=%f", s, s.current, s.last, gradient)

	if gradient == .0 {
		return false
	}

	if s.MaxGradient > 0.0 && gradient < s.MaxGradient && gradient > s.MinGradient {
		return true
	}

	return false
}
