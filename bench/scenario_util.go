package bench

import (
	"fmt"
	"log"
	"sync/atomic"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/score"
	"github.com/k0kubun/pp/v3"
)

// どのシナリオから加算されたスコアかをカウントしならがスコアを追加する
type ScenarioTag string

func (sc *Scenario) AddScoreByScenario(step *isucandar.BenchmarkStep, scoreTag score.ScoreTag, scenarioTag ScenarioTag) {
	key := fmt.Sprintf("%s", scenarioTag)
	value, ok := sc.ScenarioScoreMap.Load(key)
	if ok {
		if ptr, ok := value.(*int64); ok {
			atomic.AddInt64(ptr, ResultScoreMap[scoreTag])
		} else {
			log.Printf("error failed ScenarioScoreMap.Load type assertion: key(%s)\n", key)
		}
	} else {
		n := ResultScoreMap[scoreTag]
		sc.ScenarioScoreMap.Store(key, &n)
	}
	step.AddScore(scoreTag)
}

// シナリオ毎のスコア表示
func (sc *Scenario) PrintScenarioScoreMap() {
	ssmap := map[string]int64{}
	sc.ScenarioScoreMap.Range(func(key, value any) bool {
		tag, okKey := key.(string)
		scorePtr, okVal := value.(*int64)
		if !okKey || !okVal {
			log.Printf("error failed ScenarioScoreMap.Load type assertion: key(%s)\n", key)
			return false
		}
		scoreVal := atomic.LoadInt64(scorePtr)
		ssmap[string(tag)] = scoreVal

		return true
	})
	AdminLogger.Println(pp.Sprint(ssmap))
}

// シナリオ回転数
func (sc *Scenario) ScenarioStart(scTag ScenarioTag) {
	sc.ScenarioCountMutex.Lock()
	defer sc.ScenarioCountMutex.Unlock()

	if _, ok := sc.ScenarioCountMap[scTag]; !ok {
		sc.ScenarioCountMap[scTag] = []int{0, 0}
	}
	sc.ScenarioCountMap[scTag][0] += 1
}

func (sc *Scenario) ScenarioError(scTag ScenarioTag, err error) {
	// ctx.Doneで切られたなら何もしない
	if ve, ok := err.(ValidationError); ok && ve.Canceled {
		return
	}

	AdminLogger.Printf("[%s]: %s", scTag, err)
	sc.ScenarioCountMutex.Lock()
	defer sc.ScenarioCountMutex.Unlock()

	if _, ok := sc.ScenarioCountMap[scTag]; !ok {
		sc.ScenarioCountMap[scTag] = []int{0, 0}
	}
	sc.ScenarioCountMap[scTag][1] += 1
}

func (sc *Scenario) PrintScenarioCount() {
	sc.ScenarioCountMutex.Lock()
	defer sc.ScenarioCountMutex.Unlock()
	scmap := map[ScenarioTag]string{}
	for key, value := range sc.ScenarioCountMap {
		scmap[key] = fmt.Sprintf("count: %d (error: %d)", value[0], value[1])
	}
	AdminLogger.Println(pp.Sprint(scmap))
}