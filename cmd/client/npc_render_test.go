//go:build noebiten

// Test file for NPC rendering logic (noebiten build tag for CI).
package main

import (
	"testing"

	"github.com/opd-ai/wyrm/pkg/rendering/sprite"
)

func TestNewNPCRenderer(t *testing.T) {
	r := NewNPCRenderer("fantasy", 12345)
	if r == nil {
		t.Fatal("NewNPCRenderer returned nil")
	}
	if r.genre != "fantasy" {
		t.Errorf("expected genre 'fantasy', got %q", r.genre)
	}
	if r.seed != 12345 {
		t.Errorf("expected seed 12345, got %d", r.seed)
	}
	if r.generator == nil {
		t.Error("generator is nil")
	}
	if r.cache == nil {
		t.Error("cache is nil")
	}
}

func TestActivityToAnimState(t *testing.T) {
	tests := []struct {
		activity string
		want     string
	}{
		{"sleep", sprite.AnimDead},
		{"rest", sprite.AnimDead},
		{"walk", sprite.AnimWalk},
		{"travel", sprite.AnimWalk},
		{"patrol", sprite.AnimWalk},
		{"run", sprite.AnimRun},
		{"flee", sprite.AnimRun},
		{"chase", sprite.AnimRun},
		{"work", sprite.AnimWork},
		{"craft", sprite.AnimWork},
		{"trade", sprite.AnimWork},
		{"eat", sprite.AnimSit},
		{"drink", sprite.AnimSit},
		{"socialize", sprite.AnimIdle},
		{"talk", sprite.AnimIdle},
		{"fight", sprite.AnimAttack},
		{"attack", sprite.AnimAttack},
		{"sneak", sprite.AnimSneak},
		{"hide", sprite.AnimSneak},
		{"cast", sprite.AnimCast},
		{"magic", sprite.AnimCast},
		{"unknown", sprite.AnimIdle}, // default
		{"", sprite.AnimIdle},        // empty string default
	}

	for _, tt := range tests {
		t.Run(tt.activity, func(t *testing.T) {
			got := activityToAnimState(tt.activity)
			if got != tt.want {
				t.Errorf("activityToAnimState(%q) = %q, want %q", tt.activity, got, tt.want)
			}
		})
	}
}

func TestNPCRendererGetAnimationFrameCount(t *testing.T) {
	r := NewNPCRenderer("fantasy", 12345)

	tests := []struct {
		animState string
		want      int
	}{
		{sprite.AnimIdle, 4},
		{sprite.AnimWalk, 8},
		{sprite.AnimRun, 8},
		{sprite.AnimAttack, 6},
		{sprite.AnimCast, 8},
		{sprite.AnimSneak, 8},
		{sprite.AnimDead, 1},
		{sprite.AnimSit, 1},
		{sprite.AnimWork, 4},
		{"unknown", 4}, // default
	}

	for _, tt := range tests {
		t.Run(tt.animState, func(t *testing.T) {
			app := &mockAppearance{animState: tt.animState}
			got := r.getAnimationFrameCount(app)
			if got != tt.want {
				t.Errorf("getAnimationFrameCount(%q) = %d, want %d", tt.animState, got, tt.want)
			}
		})
	}
}

func TestNPCRendererCacheStats(t *testing.T) {
	r := NewNPCRenderer("fantasy", 12345)

	hits, misses, evicts := r.CacheStats()
	if hits != 0 || misses != 0 || evicts != 0 {
		t.Errorf("expected empty cache stats, got hits=%d, misses=%d, evicts=%d", hits, misses, evicts)
	}

	size := r.CacheSize()
	if size != 0 {
		t.Errorf("expected cache size 0, got %d", size)
	}

	mem := r.CacheMemory()
	if mem != 0 {
		t.Errorf("expected cache memory 0, got %d", mem)
	}
}
