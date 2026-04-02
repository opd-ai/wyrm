package systems

import (
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// Gossip system constants.
const (
	// DefaultGossipChance is the default probability of sharing gossip.
	DefaultGossipChance = 0.3
	// DefaultListenChance is the default probability of remembering gossip.
	DefaultListenChance = 0.7
	// DefaultGossipCooldown is the minimum seconds between gossip sharing.
	DefaultGossipCooldown = 60.0
	// GossipDecayRate is how fast gossip spreads per update (per second).
	GossipDecayRate = 0.001
	// MaxGossipAge is the maximum age of gossip before it's forgotten (seconds).
	MaxGossipAge = 86400.0 // 24 hours
	// GossipProximityRange is the range for NPCs to share gossip.
	GossipProximityRange = 10.0
)

// GossipSystem handles gossip propagation between NPCs.
type GossipSystem struct {
	// GameTime tracks elapsed time.
	GameTime float64
}

// NewGossipSystem creates a new gossip system.
func NewGossipSystem() *GossipSystem {
	return &GossipSystem{GameTime: 0}
}

// Update processes gossip decay and propagation.
func (s *GossipSystem) Update(w *ecs.World, dt float64) {
	s.GameTime += dt
	s.decayOldGossip(w)
	s.propagateGossip(w, dt)
}

// decayOldGossip removes gossip that's too old.
func (s *GossipSystem) decayOldGossip(w *ecs.World) {
	for _, e := range w.Entities("GossipNetwork") {
		gossipComp, ok := w.GetComponent(e, "GossipNetwork")
		if !ok {
			continue
		}
		gossip := gossipComp.(*components.GossipNetwork)

		for id, item := range gossip.KnownGossip {
			if s.GameTime-item.OriginTime > MaxGossipAge {
				delete(gossip.KnownGossip, id)
			}
		}
	}
}

// propagateGossip spreads gossip between nearby NPCs.
func (s *GossipSystem) propagateGossip(w *ecs.World, dt float64) {
	entities := w.Entities("GossipNetwork", "Position")
	for i, e1 := range entities {
		if !s.canShareGossip(w, e1) {
			continue
		}
		s.tryShareWithNearbyNPCs(w, e1, entities, i, dt)
	}
}

// canShareGossip checks if an entity is off cooldown for sharing.
func (s *GossipSystem) canShareGossip(w *ecs.World, e ecs.Entity) bool {
	g1Comp, ok := w.GetComponent(e, "GossipNetwork")
	if !ok {
		return false
	}
	g1 := g1Comp.(*components.GossipNetwork)
	return s.GameTime-g1.LastGossipTime >= g1.GossipCooldown
}

// tryShareWithNearbyNPCs attempts to share gossip with all nearby NPCs.
func (s *GossipSystem) tryShareWithNearbyNPCs(w *ecs.World, e1 ecs.Entity, entities []ecs.Entity, i int, dt float64) {
	g1Comp, ok := w.GetComponent(e1, "GossipNetwork")
	if !ok {
		return
	}
	g1 := g1Comp.(*components.GossipNetwork)
	p1Comp, ok := w.GetComponent(e1, "Position")
	if !ok {
		return
	}
	p1 := p1Comp.(*components.Position)

	for j, e2 := range entities {
		if i >= j {
			continue // Skip self and already-checked pairs
		}
		s.tryExchangeGossipWithNPC(w, g1, p1, e2, dt)
	}
}

// tryExchangeGossipWithNPC exchanges gossip with a single NPC if in range.
func (s *GossipSystem) tryExchangeGossipWithNPC(w *ecs.World, g1 *components.GossipNetwork, p1 *components.Position, e2 ecs.Entity, dt float64) {
	p2Comp, ok := w.GetComponent(e2, "Position")
	if !ok {
		return
	}
	p2 := p2Comp.(*components.Position)

	if !s.inGossipRange(p1, p2) {
		return
	}

	g2Comp, ok := w.GetComponent(e2, "GossipNetwork")
	if !ok {
		return
	}
	g2 := g2Comp.(*components.GossipNetwork)
	s.exchangeGossip(g1, g2, dt)
}

// inGossipRange checks if two positions are close enough for gossip.
func (s *GossipSystem) inGossipRange(p1, p2 *components.Position) bool {
	dx := p1.X - p2.X
	dy := p1.Y - p2.Y
	dz := p1.Z - p2.Z
	distSq := dx*dx + dy*dy + dz*dz
	return distSq <= GossipProximityRange*GossipProximityRange
}

// exchangeGossip allows two NPCs to share gossip.
func (s *GossipSystem) exchangeGossip(g1, g2 *components.GossipNetwork, dt float64) {
	// g1 shares with g2
	s.shareGossipOneWay(g1, g2, dt)
	// g2 shares with g1
	s.shareGossipOneWay(g2, g1, dt)
}

// shareGossipOneWay transfers gossip from source to target.
func (s *GossipSystem) shareGossipOneWay(source, target *components.GossipNetwork, dt float64) {
	if len(source.KnownGossip) == 0 {
		return
	}
	s.ensureGossipMap(target)

	shareChance := source.GossipChance * dt
	listenChance := target.ListenChance

	for id, item := range source.KnownGossip {
		if s.shouldTransferGossip(target, id, shareChance, listenChance) {
			s.transferGossipItem(source, target, id, item, dt)
		}
	}
}

// ensureGossipMap initializes the gossip map if needed.
func (s *GossipSystem) ensureGossipMap(network *components.GossipNetwork) {
	if network.KnownGossip == nil {
		network.KnownGossip = make(map[string]*components.GossipItem)
	}
}

// shouldTransferGossip checks if gossip should be transferred.
func (s *GossipSystem) shouldTransferGossip(target *components.GossipNetwork, id string, shareChance, listenChance float64) bool {
	if _, known := target.KnownGossip[id]; known {
		return false
	}
	return shareChance > 0 && listenChance > 0
}

// transferGossipItem copies gossip from source to target.
func (s *GossipSystem) transferGossipItem(source, target *components.GossipNetwork, id string, item *components.GossipItem, dt float64) {
	target.KnownGossip[id] = &components.GossipItem{
		ID:                 item.ID,
		Topic:              item.Topic,
		Content:            item.Content,
		SubjectEntity:      item.SubjectEntity,
		OriginTime:         item.OriginTime,
		Spread:             item.Spread + GossipDecayRate*dt,
		Truthfulness:       item.Truthfulness,
		ImpactOnReputation: item.ImpactOnReputation,
	}
	item.Spread += GossipDecayRate * dt
	if item.Spread > 1.0 {
		item.Spread = 1.0
	}
	source.LastGossipTime = s.GameTime
}

// CreateGossip generates a new piece of gossip.
func (s *GossipSystem) CreateGossip(w *ecs.World, npc ecs.Entity, id, topic, content string, subject ecs.Entity, truthfulness float64) bool {
	gossipComp, ok := w.GetComponent(npc, "GossipNetwork")
	if !ok {
		return false
	}
	gossip := gossipComp.(*components.GossipNetwork)

	if gossip.KnownGossip == nil {
		gossip.KnownGossip = make(map[string]*components.GossipItem)
	}

	gossip.KnownGossip[id] = &components.GossipItem{
		ID:                 id,
		Topic:              topic,
		Content:            content,
		SubjectEntity:      uint64(subject),
		OriginTime:         s.GameTime,
		Spread:             0.0,
		Truthfulness:       truthfulness,
		ImpactOnReputation: s.calculateReputationImpact(topic, truthfulness),
	}
	return true
}

// calculateReputationImpact determines how gossip affects reputation.
func (s *GossipSystem) calculateReputationImpact(topic string, truthfulness float64) float64 {
	baseImpact := 0.0
	switch topic {
	case "crime":
		baseImpact = -0.2
	case "romance":
		baseImpact = -0.05 // Mild scandal
	case "business":
		baseImpact = 0.0 // Neutral unless negative
	case "politics":
		baseImpact = -0.1
	case "danger":
		baseImpact = 0.0 // Warning, not reputation
	case "heroism":
		baseImpact = 0.15
	}
	// Truthful gossip has more impact
	return baseImpact * (0.5 + truthfulness*0.5)
}

// GetGossipAbout retrieves all gossip about a specific entity.
func (s *GossipSystem) GetGossipAbout(w *ecs.World, npc, subject ecs.Entity) []*components.GossipItem {
	gossipComp, ok := w.GetComponent(npc, "GossipNetwork")
	if !ok {
		return nil
	}
	gossip := gossipComp.(*components.GossipNetwork)

	var results []*components.GossipItem
	for _, item := range gossip.KnownGossip {
		if item.SubjectEntity == uint64(subject) {
			results = append(results, item)
		}
	}
	return results
}
