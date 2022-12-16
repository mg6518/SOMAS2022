package team5

import (
	"infra/game/commons"
	"infra/game/decision"
	"infra/game/state"
	"math"
	"sort"
	"sync"

	"github.com/benbjohnson/immutable"
)

var (
	wg         sync.WaitGroup
	once       sync.Once
	totalAgent uint // 所有存在的agent
)

// FindBestStrategy ...
func FindBestStrategy(view state.View) *immutable.Map[commons.ID, decision.FightAction] {
	level := view.CurrentLevel()
	monsterHealth := view.MonsterHealth()
	monsterAttack := view.MonsterAttack()
	agentState := view.AgentState()
	agentIterator := agentState.Iterator()

	// iterator get agent value
	agents := make([]*Agent, 0)
	for !agentIterator.Done() {
		key, value, _ := agentIterator.Next()
		agent := &Agent{
			ID:      key,
			Hp:      uint(value.Hp),
			Attack:  uint(value.Attack + value.BonusAttack),
			Defense: uint(value.Defense + value.BonusDefense),
		}
		agents = append(agents, agent)
	}

	// 所有人
	once.Do(func() {
		totalAgent = uint(len(agents))
	})
	nowAgents := uint(len(agents))

	// get allAttack & allDefence
	allAttack := uint(0)
	allDefence := uint(0)
	for _, item := range agents {
		allAttack += item.Attack
		allDefence += item.Defense
	}

	// allAttack > monsterHealth ,success
	if allAttack > monsterHealth {
		// 根据攻击排序
		sort.Slice(agents, func(i, j int) bool {
			return agents[i].Attack > agents[j].Attack
		})
		sumAttack := 0
		var index int
		// 找出总攻击>怪物血量的那个人的位置
		for i, item := range agents {
			sumAttack += int(item.Attack)
			// 全部攻击
			agents[i].Action = uint(decision.Attack)
			// 找到指定位置的人了
			if sumAttack >= int(monsterHealth) {
				index = i
				break
			}
		}
		// 如果不是最后一个人
		for i, _ := range agents {
			if i <= index {
				continue
			}
			// 指定位置之后的人都逃跑
			agents[i].Action = uint(decision.Cower)
		}
		return ConvertToImmutable(agents, agents)
	}

	// allDefence > monsterAttack, success
	if allDefence > monsterAttack {
		// 根据防御排序
		sort.Slice(agents, func(i, j int) bool {
			return agents[i].Defense > agents[j].Defense
		})
		sumDefence := 0
		var index int
		// 找出总防御>怪物攻击的那个人的位置
		for i, item := range agents {
			sumDefence += int(item.Defense)
			// 全部防御
			agents[i].Action = uint(decision.Defend)
			// 找到指定位置的人了
			if sumDefence >= int(monsterAttack) {
				index = i
				break
			}
		}
		// 如果不是最后一个人
		if index < len(agents)-1 {
			for i, _ := range agents {
				if i <= index {
					continue
				}
				// 指定位置之后的人都攻击
				agents[i].Action = uint(decision.Attack)
			}
			return ConvertToImmutable(agents, agents)
		}
	}

	// sort by Hp
	sort.Slice(agents, func(i, j int) bool {
		return agents[i].Hp > agents[j].Hp
	})

	// build new agent (weight attack 0.5,0.75,1,1.5,2)
	agent1 := BuildNewAgent(agents, 0.5)
	agent2 := BuildNewAgent(agents, 0.75)
	agent3 := BuildNewAgent(agents, 1)
	agent4 := BuildNewAgent(agents, 1.5)
	agent5 := BuildNewAgent(agents, 2)

	// handle service
	result1 := make([]*Result, len(agent1))
	for i := 1; i <= len(agent1); i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			result1[i-1] = handleStrategyService(monsterHealth, monsterAttack, agents[0:i])
		}(i)
	}
	wg.Wait()

	// handle service
	result2 := make([]*Result, len(agent2))
	for i := 1; i <= len(agent2); i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			result2[i-1] = handleStrategyService(monsterHealth, monsterAttack, agents[0:i])
		}(i)
	}
	wg.Wait()

	// handle service
	result3 := make([]*Result, len(agent3))
	for i := 1; i <= len(agent3); i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			result3[i-1] = handleStrategyService(monsterHealth, monsterAttack, agents[0:i])
		}(i)
	}
	wg.Wait()

	// handle service
	result4 := make([]*Result, len(agent4))
	for i := 1; i <= len(agent4); i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			result4[i-1] = handleStrategyService(monsterHealth, monsterAttack, agents[0:i])
		}(i)
	}
	wg.Wait()

	// handle service
	result5 := make([]*Result, len(agent5))
	for i := 1; i <= len(agent5); i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			result5[i-1] = handleStrategyService(monsterHealth, monsterAttack, agents[0:i])
		}(i)
	}
	wg.Wait()

	resultAll := make([]*Result, 0)
	resultAll = append(resultAll, result1...)
	resultAll = append(resultAll, result2...)
	resultAll = append(resultAll, result3...)
	resultAll = append(resultAll, result4...)
	resultAll = append(resultAll, result5...)

	minDeathRes := &Result{Death: math.MaxInt}
	minDamageRes := &Result{Damage: math.MaxUint}

	// 死亡人数和总人数不等时，找出最少死亡的
	// 找出最少伤害的
	suviveRate := float64(nowAgents) / float64(totalAgent)
	population := GetPopulation(level)
	if suviveRate >= population {
		for _, item := range resultAll {
			if item.Death != len(item.Agents) && item.Damage < minDamageRes.Damage {
				minDamageRes = item
			}
		}
		// 有符合条件的
		if len(minDamageRes.Agents) > 0 {
			return ConvertToImmutable(minDamageRes.Agents, agents)
		}
	} else {
		for _, item := range resultAll {
			if item.Death != len(item.Agents) && uint(item.Death) < uint(minDeathRes.Death) {
				minDeathRes = item
			}
		}
		// 有符合条件的
		if len(minDeathRes.Agents) > 0 {
			return ConvertToImmutable(minDeathRes.Agents, agents)
		}
	}
	return nil
}

func handleStrategyService(monsterHealthy, monsterAttack uint, agents []*Agent) *Result {
	newAgents := CopySlice(agents)
	totalAttack := uint(0)
	totalDefense := uint(0)
	for _, item := range newAgents {
		totalAttack += item.Attack
		totalDefense += item.Defense
	}

	// total rounds
	rounds := monsterHealthy / totalAttack
	if monsterHealthy%totalAttack != 0 {
		rounds += 1
	}

	// total Damage
	totalDamage := (monsterAttack - totalDefense) * rounds
	divideDamage := totalDamage / uint(len(newAgents))
	death := 0
	for _, item := range newAgents {
		if divideDamage >= item.Hp {
			death++
		}
	}
	res := &Result{
		Damage: divideDamage,
		Death:  death,
		Agents: agents,
	}
	return res
}