/*******************************************************
* Copyright (C) 2022 Team 1 @ SOMAS2022
*
* This file is part of SOMAS2022.
*
* This file or its contents can not be copied and/or used
* without the express permission of Team 1, SOMAS2022
*******************************************************/

package team1

import (
	"infra/game/agent"
)

// This type will make it easier to extract from map, sort, and retrieve agent ID
type SocialCapInfo struct {
	ID  string
	arr [4]float64
}

// Demonstrate creating a strategy with input parameters
func CreateCollaborativeAgent() agent.Strategy {
	return NewSocialAgent(0.7)
}

func CreateSelfishAgent() agent.Strategy {
	return NewSocialAgent(0.3)
}