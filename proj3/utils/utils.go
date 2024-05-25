package utils

import (
	"log"
	"math"
	"math/rand"
	"sync/atomic"

	"gonum.org/v1/gonum/stat/distuv"
)

type Request struct {
	SimulationSize             int     `json:"size"`
	Symbol                     string  `json:"symbol"`
	HypothesisPercentageChange float64 `json:"percentage_change"`
	HypothesisDaysAhead        int     `json:"in_x_days"`
}

type Task struct {
	NormalDist distuv.Normal
	Request
	NorminvToCumulative chan float64
	NTCChannelOpen      int64        //this will be 0 when open, 1 when closed
	CumulativeToPool    chan float64 //Let this have capacity equal to length of sample
}

type CompletedTask struct {
	Probability float64 `json:"probability"`
	RequestInfo Request `json:"request_info"`
}

func (r Request) RequestToTask(normdist *distuv.Normal) *Task {
	newtask := Task{NormalDist: *normdist, Request: r, NorminvToCumulative: make(chan float64, r.SimulationSize), CumulativeToPool: make(chan float64, r.SimulationSize)}
	return &newtask
}

// Given a channel that streams data, returns the mean and stdv
func GetStats(ch <-chan float64, totalsum float64, totallen int) (mean float64, stdv float64) {
	log.Println("total len", totallen)
	mean = totalsum / float64(totallen)
	var sumsquares, val float64
	for i := 0; i < totallen; i++ {
		val = <-ch
		sumsquares += (val - mean) * (val - mean)
	}

	variance := float64(sumsquares / (float64(totallen)))
	// log.Printf("sumsquares: %.20f\n", sumsquares)
	// log.Printf("variance %.20f", variance)
	stdv = (math.Sqrt(variance))
	// log.Printf("stdv %.20f", stdv)
	return mean, stdv
}

// Using the Normal Distribution approximated for each stock, we run hypothetical values for the dailymovement for each day. Ultimately we feed this into the ConsolidateCumulative function to get the cumulative change over the timeframe we are interested in.
func GetNormInv(task *Task) {
	var cumulative, daily float64
	for atomic.LoadInt64(&task.NTCChannelOpen) == 0 {
		cumulative = 1
		for day := 0; day < task.HypothesisDaysAhead; day++ {
			randomNumber := rand.Float64()
			// log.Printf("	this is random number: %f this is norminverse: %f\n", randomNumber, task.NormalDist.Quantile(randomNumber))
			daily = task.NormalDist.Quantile(randomNumber) / 100
			cumulative *= (1 + daily)
		}
		task.NorminvToCumulative <- cumulative - 1
	}
}

// We generate large number of hypothetical values for the cumulative change during the timeframe we are interested in. Once we have a large dataset with these cumulative values, we can again construct a normal distribution to see the likeliness of the original hypothesis. This function will not return any value as the computation will be sent to GetStats directly.
func ConsolidateCumulative(task *Task) (mean float64, stdv float64) {
	var totalsum float64
	for i := 0; i < task.SimulationSize; i++ {
		//similar to how api call is processed, first want to make a run and store the totalsum, then also persist the channel into another channel
		val := <-task.NorminvToCumulative
		task.CumulativeToPool <- val
		totalsum += val
		// log.Println("	CUM->", val)
	}
	atomic.CompareAndSwapInt64(&task.NTCChannelOpen, 0, 1)
	close(task.CumulativeToPool)
	return GetStats(task.CumulativeToPool, totalsum, task.SimulationSize)
}

func GetHypothesisProbability(mean, stdv, hval float64) float64 {
	mean *= 100
	stdv *= 100
	normaldist := distuv.Normal{
		Mu:    mean,
		Sigma: stdv,
	}
	hval /= 100
	log.Println("hval: ", hval)
	probability := 1 - normaldist.CDF(hval)
	return probability
}
