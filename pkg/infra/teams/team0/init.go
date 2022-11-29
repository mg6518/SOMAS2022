package team0

import (
	"infra/config"
	"infra/game/agent"
	"infra/game/commons"
	"infra/game/example"
	"infra/game/stage/initialise"
	"infra/game/state"
)

/**
 * This is an example of a private experiment:
 *
 * Try running this several times (set `MODE=0` in .env) and observe the final levels reached.
 * Now uncomment the "DefensiveAgent" amd comment out the "AggressiveAgent", what differences
 * do you observe?
 */
var InitAgentMap = map[commons.ID]func() agent.Strategy{
	"RANDOM":          example.NewRandomAgent,
	"AggressiveAgent": CreateAggressiveAgent,
	"DefensiveAgent":  CreateDefensiveAgent,
	// "CowardlyAgent": CreateCoweringAgent,
}

func InitAgents(defaultStrategyMap map[commons.ID]func() agent.Strategy, gameConfig config.GameConfig) (numAgents uint, agentMap map[commons.ID]agent.Agent, agentStateMap map[commons.ID]state.AgentState) {
	agentMap = make(map[commons.ID]agent.Agent)
	agentStateMap = make(map[commons.ID]state.AgentState)

	numAgents = 0

	for agentName, strategy := range defaultStrategyMap {
		expectedEnvName := "AGENT_" + agentName + "_QUANTITY"
		quantity := config.EnvToUint(expectedEnvName, 100)

		numAgents += quantity
		initialise.InstantiateAgent(gameConfig, agentMap, agentStateMap, quantity, strategy, agentName)
	}

	return
}
