package team3

import (
	"math/rand"

	"infra/game/agent"
	"infra/game/commons"
	"infra/game/decision"
	"infra/game/message"
	"infra/game/message/proposal"
	"infra/game/state"
	"infra/logging"

	"github.com/benbjohnson/immutable"
)

const PERCENTAGE = 500

type Utility struct {
	ID    commons.ID
	score int
}

type UtilityMap []Utility

func (u UtilityMap) Len() int           { return len(u) }
func (u UtilityMap) Swap(i, j int)      { u[i], u[j] = u[j], u[i] }
func (u UtilityMap) Less(i, j int) bool { return u[i].score < u[j].score }

type AgentThree struct {
	HP           int
	ST           int
	AT           int
	bravery      int
	utilityScore map[commons.ID]int
}

// HP pool donation
func (a *AgentThree) DonateToHpPool(baseAgent agent.BaseAgent) uint {
	donation := rand.Intn(2)
	// If our health is > 50% and we feel generous then donate some (max 20%) HP
	if donation == 1 && a.HP > PERCENTAGE {
		return uint(rand.Intn((a.HP * 20) / 100))
	} else {
		return 0
	}
}

// Update internal parameters at the end of each lvl!?
func (a *AgentThree) UpdateInternalState(baseAgent agent.BaseAgent, _ *commons.ImmutableList[decision.ImmutableFightResult], _ *immutable.Map[decision.Intent, uint]) {
	a.UpdateUtility(baseAgent)
	a.HP = int(baseAgent.AgentState().Hp)
	a.ST = int(baseAgent.AgentState().Stamina)
	a.AT = int(baseAgent.AgentState().Attack)
}

// Update Utility
func (a *AgentThree) UpdateUtility(baseAgent agent.BaseAgent) {
	view := baseAgent.View()
	agentState := view.AgentState()
	itr := agentState.Iterator()
	for !itr.Done() {
		id, _, ok := itr.Next()
		if !ok {
			break
		}

		// We are already cool, don't need the utility score for ourselves
		if id != baseAgent.ID() {
			a.utilityScore[id] = rand.Intn(10)
		}
	}
	// Sort utility map
	// sort.Sort(a.utilityScore)
}

func (a *AgentThree) LootAction() immutable.List[commons.ItemID] {
	return *immutable.NewList[commons.ItemID]()
}

func (a *AgentThree) FightAction(baseAgent agent.BaseAgent) decision.FightAction {
	fight := rand.Intn(3)
	switch fight {
	case 0:
		return decision.Cower
	case 1:
		return decision.Attack
	default:
		return decision.Defend
	}
}

// Create proposal for fight decisions
// func (a *AgentThree) FightResolution(baseAgent agent.BaseAgent) message.MapProposal[decision.FightAction] {
// 	actions := make(map[commons.ID]decision.FightAction)
// 	view := baseAgent.View()
// 	agentState := view.AgentState()
// 	itr := agentState.Iterator()
// 	for !itr.Done() {
// 		id, _, ok := itr.Next()
// 		if !ok {
// 			break
// 		}

// 		// Check for our agent and assign what we want to do
// 		if id == baseAgent.ID() {
// 			actions[id] = a.CurrentAction()
// 			baseAgent.Log(logging.Trace, logging.LogField{"bravery": a.bravery, "hp": a.HP, "choice": a.CurrentAction(), "util": a.utilityScore[view.CurrentLeader()]}, "Intent")
// 		} else {
// 			// Send some messages to other agents
// 			// send := rand.Intn(5)
// 			// if send == 0 {
// 			// 	m := message.FightInform()
// 			// 	_ = baseAgent.SendBlockingMessage(id, m)
// 			// }
// 			rNum := rand.Intn(3)
// 			switch rNum {
// 			case 0:
// 				actions[id] = decision.Attack
// 			case 1:
// 				actions[id] = decision.Defend
// 			default:
// 				actions[id] = decision.Cower
// 			}
// 		}
// 	}

// 	prop := message.NewProposal(uuid.NewString(), commons.MapToImmutable(actions))
// 	return *prop
// }

func (a *AgentThree) FightResolution(_ agent.BaseAgent) commons.ImmutableList[proposal.Rule[decision.FightAction]] {
	rules := make([]proposal.Rule[decision.FightAction], 0)

	rules = append(rules, *proposal.NewRule[decision.FightAction](decision.Attack,
		proposal.NewAndCondition(*proposal.NewComparativeCondition(proposal.Health, proposal.GreaterThan, 1000),
			*proposal.NewComparativeCondition(proposal.Stamina, proposal.GreaterThan, 1000)),
	))

	rules = append(rules, *proposal.NewRule[decision.FightAction](decision.Defend,
		proposal.NewComparativeCondition(proposal.TotalDefence, proposal.GreaterThan, 1000),
	))

	rules = append(rules, *proposal.NewRule[decision.FightAction](decision.Cower,
		proposal.NewComparativeCondition(proposal.Health, proposal.LessThan, 1),
	))

	rules = append(rules, *proposal.NewRule[decision.FightAction](decision.Attack,
		proposal.NewComparativeCondition(proposal.Stamina, proposal.GreaterThan, 10),
	))

	return *commons.NewImmutableList(rules)
}

// Manifesto
func (a *AgentThree) CreateManifesto(_ agent.BaseAgent) *decision.Manifesto {
	manifesto := decision.NewManifesto(false, false, 10, 50)
	return manifesto
}

// Handle No Confidence vote
func (a *AgentThree) HandleConfidencePoll(baseAgent agent.BaseAgent) decision.Intent {
	view := baseAgent.View()
	// Vote for leader to stay if he's our friend :)
	if a.utilityScore[view.CurrentLeader()] > 5 {
		return decision.Positive
	} else {
		switch rand.Intn(2) {
		case 0:
			return decision.Abstain
		default:
			return decision.Negative
		}
	}
}

// Send proposal to leader
func (a *AgentThree) HandleFightInformation(_ message.TaggedInformMessage[message.FightInform], baseAgent agent.BaseAgent, log *immutable.Map[commons.ID, decision.FightAction]) {
	// baseAgent.Log(logging.Trace, logging.LogField{"bravery": r.bravery, "hp": baseAgent.AgentState().Hp}, "Cowering")
	makesProposal := rand.Intn(100)

	// Well, not everytime. Just sometimes
	if makesProposal > 80 {
		prop := a.FightResolution(baseAgent)
		_ = baseAgent.SendFightProposalToLeader(prop)
	}
}

// Calculate our agents action
func (a *AgentThree) CurrentAction() decision.FightAction {
	// Always check how bravee we are and if we have the Stamina to do it...
	if a.bravery > 3 && a.ST > PERCENTAGE {
		fight := rand.Intn(2)
		switch fight {
		case 0:
			return decision.Attack
		default:
			return decision.Defend
		}
	} else {
		return decision.Cower
	}
}

func (a *AgentThree) HandleElectionBallot(baseAgent agent.BaseAgent, _ *decision.ElectionParams) decision.Ballot {
	// Extract ID of alive agents
	view := baseAgent.View()
	agentState := view.AgentState()
	aliveAgentIDs := make([]string, agentState.Len())
	i := 0
	itr := agentState.Iterator()
	for !itr.Done() {
		id, a, ok := itr.Next()
		if ok && a.Hp > 0 {
			aliveAgentIDs[i] = id
			i++
		}
	}

	// Randomly fill the ballot
	var ballot decision.Ballot
	numAliveAgents := len(aliveAgentIDs)
	numCandidate := 2
	for i := 0; i < numCandidate; i++ {
		randomIdx := rand.Intn(numAliveAgents)
		randomCandidate := aliveAgentIDs[uint(randomIdx)]
		ballot = append(ballot, randomCandidate)
	}

	return ballot
}

// Vote on proposal
func (a *AgentThree) HandleFightProposal(m message.Proposal[decision.FightAction], baseAgent agent.BaseAgent) decision.Intent {
	agree := true

	rules := m.Rules()
	itr := rules.Iterator()
	for !itr.Done() {
		rule, _ := itr.Next()
		baseAgent.Log(logging.Trace, logging.LogField{"rule": rule}, "Rule Proposal")
	}

	// Selfish, only agree if our decision is ok
	if agree {
		return decision.Positive
	} else {
		return decision.Negative
	}
}

func (a *AgentThree) HandleUpdateWeapon(baseAgent agent.BaseAgent) decision.ItemIdx {
	// weapons := b.AgentState().Weapons
	// return decision.ItemIdx(rand.Intn(weapons.Len() + 1))

	// 0th weapon has greatest attack points
	return decision.ItemIdx(0)
}

func (a *AgentThree) HandleUpdateShield(baseAgent agent.BaseAgent) decision.ItemIdx {
	// shields := b.AgentState().Shields
	// return decision.ItemIdx(rand.Intn(shields.Len() + 1))
	return decision.ItemIdx(0)
}

// Leader function to grant the floor
func (a *AgentThree) HandleFightProposalRequest(_ message.Proposal[decision.FightAction], _ agent.BaseAgent, _ *immutable.Map[commons.ID, decision.FightAction]) bool {
	switch rand.Intn(2) {
	case 0:
		return true
	default:
		return false
	}
}

func (a *AgentThree) HandleFightRequest(_ message.TaggedRequestMessage[message.FightRequest], _ *immutable.Map[commons.ID, decision.FightAction]) message.FightInform {
	return nil
}

func (a *AgentThree) HandleLootInformation(m message.TaggedInformMessage[message.LootInform], agent agent.BaseAgent) {
}

func (a *AgentThree) HandleLootRequest(m message.TaggedRequestMessage[message.LootRequest]) message.LootInform {
	//TODO implement me
	panic("implement me")
}

func (a *AgentThree) HandleLootProposal(_ message.Proposal[decision.LootAction], _ agent.BaseAgent) decision.Intent {
	switch rand.Intn(3) {
	case 0:
		return decision.Positive
	case 1:
		return decision.Negative
	default:
		return decision.Abstain
	}
}

func (a *AgentThree) HandleLootProposalRequest(_ message.Proposal[decision.LootAction], _ agent.BaseAgent) bool {
	switch rand.Intn(2) {
	case 0:
		return true
	default:
		return false
	}
}

func (a *AgentThree) LootAllocation(ba agent.BaseAgent) immutable.Map[commons.ID, immutable.List[commons.ItemID]] {
	lootAllocation := make(map[commons.ID][]commons.ItemID)
	view := ba.View()
	ids := commons.ImmutableMapKeys(view.AgentState())
	iterator := ba.Loot().Weapons().Iterator()
	allocateRandomly(iterator, ids, lootAllocation)
	iterator = ba.Loot().Shields().Iterator()
	allocateRandomly(iterator, ids, lootAllocation)
	iterator = ba.Loot().HpPotions().Iterator()
	allocateRandomly(iterator, ids, lootAllocation)
	iterator = ba.Loot().StaminaPotions().Iterator()
	allocateRandomly(iterator, ids, lootAllocation)
	mMapped := make(map[commons.ID]immutable.List[commons.ItemID])
	for id, itemIDS := range lootAllocation {
		mMapped[id] = commons.ListToImmutable(itemIDS)
	}
	return commons.MapToImmutable(mMapped)
}

func allocateRandomly(iterator commons.Iterator[state.Item], ids []commons.ID, lootAllocation map[commons.ID][]commons.ItemID) {
	for !iterator.Done() {
		next, _ := iterator.Next()
		toBeAllocated := ids[rand.Intn(len(ids))]
		if l, ok := lootAllocation[toBeAllocated]; ok {
			l = append(l, next.Id())
			lootAllocation[toBeAllocated] = l
		} else {
			l := make([]commons.ItemID, 0)
			l = append(l, next.Id())
			lootAllocation[toBeAllocated] = l
		}
	}
}

func NewAgentThree() agent.Strategy {
	return &AgentThree{
		bravery:      rand.Intn(10),
		utilityScore: make(map[string]int),
	}
}
