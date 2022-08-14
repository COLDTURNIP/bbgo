package kfzlema

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	//"github.com/wcharczuk/go-chart/v2"

	"github.com/c9s/bbgo/pkg/bbgo"
	"github.com/c9s/bbgo/pkg/fixedpoint"
	"github.com/c9s/bbgo/pkg/indicator"
	"github.com/c9s/bbgo/pkg/types"
)

const ID = "kfzlema"

var log = logrus.WithField("strategy", ID)

func init() {
	bbgo.RegisterStrategy(ID, &Strategy{})
}

type SourceFunc func(*types.KLine) fixedpoint.Value

type Strategy struct {
	Symbol string `json:"symbol"`

	bbgo.StrategyController
	types.Market
	types.IntervalWindow

	*bbgo.Environment
	*types.Position    `persistence:"position"`
	*types.ProfitStats `persistence:"profit_stats"`
	*types.TradeStats  `persistence:"trade_stats"`

	session     *bbgo.ExchangeSession
	signalLines *signalLines

	FilterWindow       int              `json:"filterWindow"`
	SignalWindow       int              `json:"signalWindow"`
	PriceTrackingRatio fixedpoint.Value `json:"priceTrackingRatio"`

	GraphPNLPath    string `json:"graphPNLPath"`
	GraphCumPNLPath string `json:"graphCumPNLPath"`
	// Whether to generate graph when shutdown
	GenerateGraph bool `json:"generateGraph"`

	ExitMethods bbgo.ExitMethodSet `json:"exits"`
	Session     *bbgo.ExchangeSession
	*bbgo.GeneralOrderExecutor
}

func (s *Strategy) Print(o *os.File) {
	f := bufio.NewWriter(o)
	defer f.Flush()
	b, _ := json.MarshalIndent(s.ExitMethods, "  ", "  ")
	hiyellow := color.New(color.FgHiYellow).FprintfFunc()
	hiyellow(f, "------ %s Settings ------\n", s.InstanceID())
	hiyellow(f, "generateGraph: %v\n", s.GenerateGraph)
	hiyellow(f, "graphPNLPath: %s\n", s.GraphPNLPath)
	hiyellow(f, "graphCumPNLPath: %s\n", s.GraphCumPNLPath)
	hiyellow(f, "exits:\n %s\n", string(b))
	hiyellow(f, "symbol: %s\n", s.Symbol)
	hiyellow(f, "interval: %s\n", s.Interval)
	hiyellow(f, "window: %d\n", s.Window)
	hiyellow(f, "\n")
}

func (s *Strategy) ID() string {
	return ID
}

func (s *Strategy) InstanceID() string {
	return fmt.Sprintf("%s:%s", ID, s.Symbol)
}

func (s *Strategy) Subscribe(session *bbgo.ExchangeSession) {
	session.Subscribe(types.KLineChannel, s.Symbol, types.SubscribeOptions{
		Interval: s.Interval,
	})

	s.ExitMethods.SetAndSubscribe(session, s)
}

func (s *Strategy) CurrentPosition() *types.Position {
	return s.Position
}

func (s *Strategy) ClosePosition(ctx context.Context, percentage fixedpoint.Value) error {
	_, price, _, err := s.getBidAskPrice(ctx)
	if nil != err {
		return fmt.Errorf("cannot get price spread from market")
	}

	order := s.Position.NewMarketCloseOrder(percentage)
	if order == nil {
		return nil
	}
	order.Tag = "close"
	order.TimeInForce = ""
	balances := s.Session.GetAccount().Balances()
	baseBalance := balances[s.Market.BaseCurrency].Available
	if order.Side == types.SideTypeBuy {
		quoteAmount := balances[s.Market.QuoteCurrency].Available.Div(price)
		if order.Quantity.Compare(quoteAmount) > 0 {
			order.Quantity = quoteAmount
		}
	} else if order.Side == types.SideTypeSell && order.Quantity.Compare(baseBalance) > 0 {
		order.Quantity = baseBalance
	}
	Delta := fixedpoint.NewFromFloat(0.01)
	for {
		if s.Market.IsDustQuantity(order.Quantity, price) {
			return nil
		}
		_, err := s.GeneralOrderExecutor.SubmitOrders(ctx, *order)
		if err != nil {
			order.Quantity = order.Quantity.Mul(fixedpoint.One.Sub(Delta))
			continue
		}
		return nil
	}
}

type signalLines struct {
	types.SeriesBase
	filter *indicator.KalmanFilter
	ma     *indicator.ZLEMA
	signal *indicator.ZLEMA
}

func (s *signalLines) PushK(k types.KLine) {
	s.filter.PushK(k)
	s.ma.Update(s.filter.Last())
	s.signal.Update(s.filter.Last())
}

func (s *signalLines) isLongSignal() bool {
	return s.ma.Last() > s.signal.Last()
}

func (s *signalLines) isShortSignal() bool {
	return s.ma.Last() <= s.signal.Last()
}

func (s *Strategy) initIndicators() error {
	s.signalLines = &signalLines{
		filter: &indicator.KalmanFilter{
			IntervalWindow: types.IntervalWindow{
				Interval: s.Interval,
				Window:   s.FilterWindow,
			},
		},
		ma: &indicator.ZLEMA{
			IntervalWindow: s.IntervalWindow,
		},
		signal: &indicator.ZLEMA{
			IntervalWindow: types.IntervalWindow{
				Interval: s.Interval,
				Window:   s.SignalWindow,
			},
		},
	}
	return nil
}

//func (s *Strategy) Draw(time types.Time, priceLine types.SeriesExtend, profit types.Series, cumProfit types.Series, zeroPoints types.Series) {
//	canvas := types.NewCanvas(s.InstanceID(), s.Interval)
//	Length := priceLine.Length()
//	if Length > 300 {
//		Length = 300
//	}
//	mean := priceLine.Mean(Length)
//	highestPrice := priceLine.Minus(mean).Abs().Highest(Length)
//	highestDrift := s.drift.Abs().Highest(Length)
//	hi := s.drift.drift.Abs().Highest(Length)
//	ratio := highestPrice / highestDrift
//	canvas.Plot("upband", s.ma.Add(s.stdevHigh), time, Length)
//	canvas.Plot("ma", s.ma, time, Length)
//	canvas.Plot("downband", s.ma.Minus(s.stdevLow), time, Length)
//	canvas.Plot("drift", s.drift.Mul(ratio).Add(mean), time, Length)
//	canvas.Plot("driftOrig", s.drift.drift.Mul(highestPrice/hi).Add(mean), time, Length)
//	canvas.Plot("zero", types.NumberSeries(mean), time, Length)
//	canvas.Plot("price", priceLine, time, Length)
//	canvas.Plot("zeroPoint", zeroPoints, time, Length)
//	f, err := os.Create(s.CanvasPath)
//	if err != nil {
//		log.WithError(err).Errorf("cannot create on %s", s.CanvasPath)
//		return
//	}
//	defer f.Close()
//	if err := canvas.Render(chart.PNG, f); err != nil {
//		log.WithError(err).Errorf("cannot render in drift")
//	}
//
//	canvas = types.NewCanvas(s.InstanceID())
//	f, err = os.Create(s.GraphPNLPath)
//	if err != nil {
//		log.WithError(err).Errorf("open pnl")
//		return
//	}
//	defer f.Close()
//	if err := canvas.Render(chart.PNG, f); err != nil {
//		log.WithError(err).Errorf("render pnl")
//	}
//
//	canvas = types.NewCanvas(s.InstanceID())
//	f, err = os.Create(s.GraphCumPNLPath)
//	if err != nil {
//		log.WithError(err).Errorf("open cumpnl")
//		return
//	}
//	defer f.Close()
//	if err := canvas.Render(chart.PNG, f); err != nil {
//		log.WithError(err).Errorf("render cumpnl")
//	}
//}

func (s *Strategy) Run(ctx context.Context, orderExecutor bbgo.OrderExecutor, session *bbgo.ExchangeSession) error {
	// initial required information
	s.session = session
	instanceID := s.InstanceID()

	if s.Position == nil {
		s.Position = types.NewPositionFromMarket(s.Market)
	}
	if session.MakerFeeRate.Sign() > 0 || session.TakerFeeRate.Sign() > 0 {
		s.Position.SetExchangeFeeRate(session.ExchangeName, types.ExchangeFee{
			MakerFeeRate: session.MakerFeeRate,
			TakerFeeRate: session.TakerFeeRate,
		})
	}

	if s.ProfitStats == nil {
		s.ProfitStats = types.NewProfitStats(s.Market)
	}

	if s.TradeStats == nil {
		s.TradeStats = types.NewTradeStats(s.Symbol)
	}
	startTime := s.Environment.StartTime()
	s.TradeStats.SetIntervalProfitCollector(types.NewIntervalProfitCollector(types.Interval1d, startTime))
	s.TradeStats.SetIntervalProfitCollector(types.NewIntervalProfitCollector(types.Interval1w, startTime))

	// StrategyController
	s.Status = types.StrategyStatusRunning

	s.OnSuspend(func() {
		_ = s.GeneralOrderExecutor.GracefulCancel(ctx)
	})

	s.OnEmergencyStop(func() {
		_ = s.GeneralOrderExecutor.GracefulCancel(ctx)
		_ = s.ClosePosition(ctx, fixedpoint.One)
	})

	s.GeneralOrderExecutor = bbgo.NewGeneralOrderExecutor(session, s.Symbol, ID, instanceID, s.Position)
	s.GeneralOrderExecutor.BindEnvironment(s.Environment)
	s.GeneralOrderExecutor.BindProfitStats(s.ProfitStats)
	s.GeneralOrderExecutor.BindTradeStats(s.TradeStats)
	s.GeneralOrderExecutor.TradeCollector().OnPositionUpdate(func(position *types.Position) {
		bbgo.Sync(s)
	})
	s.GeneralOrderExecutor.Bind()

	// Exit methods from config
	s.ExitMethods.Bind(session, s.GeneralOrderExecutor)

	if err := s.initIndicators(); err != nil {
		log.WithError(err).Errorf("initIndicator failed")
		return nil
	}

	dynamicKLine := &types.KLine{}
	//priceLine := types.NewQueue(300)

	session.MarketDataStream.OnKLineClosed(types.KLineWith(s.Symbol, s.Interval, func(kline types.KLine) {
		if s.Status != types.StrategyStatusRunning {
			return
		}

		if !kline.Closed {
			return
		}
		dynamicKLine.Set(&kline)

		s.signalLines.PushK(kline)
		canSell := s.signalLines.isShortSignal()
		canBuy := s.signalLines.isLongSignal()
		exitShortCondition := canBuy && s.Position.IsShort()
		exitLongCondition := canSell && s.Position.IsLong()
		bidPrice, midPrice, askPrice, err := s.getBidAskPrice(ctx)
		if nil != err {
			log.WithError(err).Errorf("cannot get price spread from market")
			return
		}

		if err := s.GeneralOrderExecutor.GracefulCancel(ctx); err != nil {
			log.WithError(err).Errorf("cannot cancel incompleted orders")
			return
		}
		if (exitShortCondition || exitLongCondition) && s.Position.IsOpened(midPrice) {
			_ = s.ClosePosition(ctx, fixedpoint.One)
			return
		}

		if canSell {
			baseBalance, ok := s.Session.GetAccount().Balance(s.Market.BaseCurrency)
			if !ok {
				log.Errorf("unable to get baseBalance")
				return
			}

			quantity := baseBalance.Available
			if s.Market.IsDustQuantity(quantity, askPrice) {
				return
			}
			if _, err := s.GeneralOrderExecutor.SubmitOrders(ctx, types.SubmitOrder{
				Symbol:   s.Symbol,
				Side:     types.SideTypeSell,
				Type:     types.OrderTypeLimit,
				Price:    askPrice,
				Quantity: quantity,
				Tag:      "short",
			}); err != nil {
				log.WithError(err).Errorf("cannot place sell order")
				return
			}
		}
		if canBuy {
			quoteBalance, ok := s.Session.GetAccount().Balance(s.Market.QuoteCurrency)
			if !ok {
				log.Errorf("unable to get quoteCurrency")
				return
			}
			quantity := quoteBalance.Available.Div(bidPrice)
			if s.Market.IsDustQuantity(quantity, bidPrice) {
				return
			}
			if _, err := s.GeneralOrderExecutor.SubmitOrders(ctx, types.SubmitOrder{
				Symbol:   s.Symbol,
				Side:     types.SideTypeBuy,
				Type:     types.OrderTypeLimit,
				Price:    bidPrice,
				Quantity: quantity,
				Tag:      "long",
			}); err != nil {
				log.WithError(err).Errorf("cannot place buy order")
				return
			}
		}
	}))

	bbgo.OnShutdown(func(ctx context.Context, wg *sync.WaitGroup) {

		defer s.Print(os.Stdout)

		defer fmt.Fprintln(os.Stdout, s.TradeStats.BriefString())

		//if s.GenerateGraph {
		//	s.Draw(dynamicKLine.StartTime, priceLine, &profit, &cumProfit, zeroPoints)
		//}

		wg.Done()
	})
	return nil
}

func (s *Strategy) getBidAskPrice(ctx context.Context) (bid, mid, ask fixedpoint.Value, err error) {
	if sBid, sAsk, err := getPriceSpread(ctx, s.session, s.Symbol); err != nil {
		return fixedpoint.Zero, fixedpoint.Zero, fixedpoint.Zero, err
	} else {
		bid = s.PriceTrackingRatio.Mul(sBid).Add(fixedpoint.One.Sub(s.PriceTrackingRatio).Mul(sAsk))
		mid = sBid.Add(sAsk).Div(fixedpoint.NewFromInt(2))
		ask = s.PriceTrackingRatio.Mul(sAsk).Add(fixedpoint.One.Sub(s.PriceTrackingRatio).Mul(sBid))
		log.Infof("using ticker price: bid %v / ask %v, calc bid %v / ask %v", sBid, sAsk, bid, ask)
		return bid, mid, ask, nil
	}
}

func getPriceSpread(ctx context.Context, session *bbgo.ExchangeSession, symbol string) (bid fixedpoint.Value, ask fixedpoint.Value, err error) {
	if bbgo.IsBackTesting {
		if price, ok := session.LastPrice(symbol); !ok {
			return fixedpoint.Zero, fixedpoint.Zero, fmt.Errorf("backtest cannot get lastprice")
		} else {
			return price, price, nil
		}
	}
	if ticker, err := session.Exchange.QueryTicker(ctx, symbol); err != nil {
		return fixedpoint.Zero, fixedpoint.Zero, err
	} else {
		return ticker.Buy, ticker.Sell, nil
	}
}
