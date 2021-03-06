// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestAudienceDisplay(t *testing.T) {
	web := setupTestWeb(t)

	recorder := web.getHttpResponse("/displays/audience")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Audience Display - Untitled Event - Cheesy Arena")
}

func TestAudienceDisplayWebsocket(t *testing.T) {
	web := setupTestWeb(t)

	server, wsUrl := web.startTestServer()
	defer server.Close()
	conn, _, err := websocket.DefaultDialer.Dial(wsUrl+"/displays/audience/websocket", nil)
	assert.Nil(t, err)
	defer conn.Close()
	ws := &Websocket{conn, new(sync.Mutex)}

	// Should get a few status updates right after connection.
	readWebsocketType(t, ws, "matchTiming")
	readWebsocketType(t, ws, "matchTime")
	readWebsocketType(t, ws, "setAudienceDisplay")
	readWebsocketType(t, ws, "setMatch")
	readWebsocketType(t, ws, "realtimeScore")
	readWebsocketType(t, ws, "setFinalScore")
	readWebsocketType(t, ws, "allianceSelection")

	// Run through a match cycle.
	web.arena.MatchLoadTeamsNotifier.Notify(nil)
	readWebsocketType(t, ws, "setMatch")
	web.arena.AllianceStations["R1"].Bypass = true
	web.arena.AllianceStations["R2"].Bypass = true
	web.arena.AllianceStations["R3"].Bypass = true
	web.arena.AllianceStations["B1"].Bypass = true
	web.arena.AllianceStations["B2"].Bypass = true
	web.arena.AllianceStations["B3"].Bypass = true
	web.arena.StartMatch()
	web.arena.Update()
	messages := readWebsocketMultiple(t, ws, 3)
	screen, ok := messages["setAudienceDisplay"]
	if assert.True(t, ok) {
		assert.Equal(t, "match", screen)
	}
	sound, ok := messages["playSound"]
	if assert.True(t, ok) {
		assert.Equal(t, "match-start", sound)
	}
	_, ok = messages["matchTime"]
	assert.True(t, ok)
	web.arena.RealtimeScoreNotifier.Notify(nil)
	readWebsocketType(t, ws, "realtimeScore")
	web.arena.ScorePostedNotifier.Notify(nil)
	readWebsocketType(t, ws, "setFinalScore")

	// Test other overlays.
	web.arena.AllianceSelectionNotifier.Notify(nil)
	readWebsocketType(t, ws, "allianceSelection")
	web.arena.LowerThirdNotifier.Notify(nil)
	readWebsocketType(t, ws, "lowerThird")
}
