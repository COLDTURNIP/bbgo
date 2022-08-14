package indicator

import (
	"github.com/c9s/bbgo/pkg/datatype/floats"
	"github.com/c9s/bbgo/pkg/types"
	"math"
)

// Refer: https://jamesgoulding.com/Research_II/Ehlers/Ehlers%20(Optimal%20Tracking%20Filters).doc
// Ehler's Optimal Tracking Filter

//go:generate callbackgen -type KalmanFilter
type KalmanFilter struct {
	types.SeriesBase
	types.IntervalWindow
	a               float64 // maneuverability uncertainty
	b               float64 // measurement uncertainty
	lastMeasurement float64
	Values          floats.Slice

	UpdateCallbacks []func(value float64)
}

func (inc *KalmanFilter) Update(value float64) {
	inc.update(value, math.Abs(value-inc.lastMeasurement))
}

func (inc *KalmanFilter) update(value, uncertainty float64) {
	if len(inc.Values) == 0 {
		inc.a = 0
		inc.b = uncertainty / 2
		inc.lastMeasurement = value
		inc.Values.Push(value)
		return
	}
	multiplier := 2.0 / float64(1+inc.Window) // EMA multiplier
	inc.a = multiplier*(value-inc.lastMeasurement) + (1-multiplier)*inc.a
	inc.b = multiplier*uncertainty/2 + (1-multiplier)*inc.b
	lambda := inc.a / inc.b
	lambda2 := lambda * lambda
	alpha := (-lambda2 + math.Sqrt(lambda2*lambda2+16*lambda2)) / 8
	filtered := alpha*value + (1-alpha)*inc.Values.Last()
	inc.Values.Push(filtered)
	inc.lastMeasurement = value
}

func (inc *KalmanFilter) Index(i int) float64 {
	if inc.Values == nil {
		return 0.0
	}
	return inc.Values.Index(i)
}

func (inc *KalmanFilter) Length() int {
	if inc.Values == nil {
		return 0
	}
	return inc.Values.Length()
}

func (inc *KalmanFilter) Last() float64 {
	if inc.Values == nil {
		return 0.0
	}
	return inc.Values.Last()
}

var _ types.SeriesExtend = &KalmanFilter{}

func (inc *KalmanFilter) PushK(k types.KLine) {
	inc.update(k.Close.Float64(), k.High.Float64()-k.Low.Float64())
}

func (inc *KalmanFilter) CalculateAndUpdate(allKLines []types.KLine) {
	if inc.Values != nil {
		k := allKLines[len(allKLines)-1]
		inc.PushK(k)
		inc.EmitUpdate(inc.Last())
		return
	}
	for _, k := range allKLines {
		inc.PushK(k)
		inc.EmitUpdate(inc.Last())
	}
}

func (inc *KalmanFilter) handleKLineWindowUpdate(interval types.Interval, window types.KLineWindow) {
	if inc.Interval != interval {
		return
	}
	inc.CalculateAndUpdate(window)
}

func (inc *KalmanFilter) Bind(updater KLineWindowUpdater) {
	updater.OnKLineWindowUpdate(inc.handleKLineWindowUpdate)
}
