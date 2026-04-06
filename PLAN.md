# Game Repair Plan

## Critical Issues (Blocks Playability)

### Rendering Pipeline
- [ ] **Renderer renders blank world when no chunk data has been received** - Location: `pkg/rendering/raycast/renderer.go:782` - Impact: `NewRendererWithGenre()` initializes `WorldMap` to a small non-nil default, so `isValidMapPosition()` safely returns false via `len(r.WorldMap)` (which is 0-safe on nil slices). However, before the first real chunk arrives, the renderer draws against this minimal default map, producing a blank/empty world with no walls. Not a crash, but a confusing first-frame experience. Consider showing a loading indicator or placeholder geometry until chunk data populates the map
- [ ] **SetWorldMapDirect leaves WorldMapCells stale when map is empty** - Location: `pkg/rendering/raycast/renderer.go:974-979` - Impact: When an empty `worldMap` is passed, `SetWorldMapDirect()` skips rebuilding `WorldMapCells` instead of clearing or regenerating it. This leaves `WorldMapCells` out of sync with `WorldMap`, so later `GetMapCell()` and raycasting logic can read stale cell data and render/collide against a previous map. Fix by explicitly clearing `WorldMapCells` when `worldMap` is empty, or by always regenerating `WorldMapCells` to match `WorldMap`
- [ ] **Framebuffer fixed at init size, no resize handling** - Location: `pkg/rendering/raycast/renderer.go:580` - Impact: Framebuffer allocated as `width*height*4` bytes at construction; if `Layout()` returns dimensions different from renderer initialization (e.g., window resize), `WritePixels` in `draw.go:20` will panic on buffer size mismatch

### Faction UI - Selection State Can Drift Out of Range
- [ ] **Faction UI selection index is only clamped in Draw, allowing invalid state between frames** - Location: `cmd/client/faction_ui.go:82-84` - Impact: `selectedFaction++` on KeyDown has no upper bound check in `Update()`; the clamp only happens later in `Draw()` at line 178-185. `adjustScroll()` itself only does arithmetic, so this is not an immediate out-of-bounds access there, but it does allow `selectedFaction` to drift beyond the available faction count between `Update()` and `Draw()` or across overlay transitions. That leaves the UI in an inconsistent state until rendering occurs and can cause bad scroll/selection behavior or later out-of-range access when code indexes using the stale selection. Clamp in `Update()` as well as `Draw()` to keep selection state valid at all times

### Crafting UI - Recipe Selection Unbounded
- [ ] **Crafting recipe selection has no upper bound** - Location: `cmd/client/crafting_ui.go:84-86` - Impact: `selectedRecipe++` on KeyDown has no max bound relative to available recipes. When `selectedRecipe` exceeds the valid range, `Draw()` and `startCraft()` guard against out-of-range access, so this does not hard-lock or crash the crafting system; instead the UI can become desynced/blank with no valid selection highlight and crafting temporarily unavailable until the player navigates back into range. Clamp `selectedRecipe` to `len(recipes)-1` in `Update()` to keep selection state valid and avoid the confusing UX

## High Priority (Degrades Experience)

### Defensive Hardening
- [ ] **Dialog greeting accesses vocab arrays without empty-check (future-proofing)** - Location: `cmd/client/dialog_ui.go:341` - Impact: `generateGreeting()` accesses `vocab.Greetings[0]` unconditionally before checking length at line 342. Currently all 5 genre vocabularies in `pkg/dialog/system.go` have non-empty `Greetings`, `Affirmatives`, and `Farewells` arrays, and `GetVocabulary()` falls back to fantasy for unknown genres, so this is not a present crash. However, if a future genre or dynamic vocabulary override provides empty arrays, this becomes an index-out-of-bounds panic. Add `len()` guards before indexing for defensive safety

### Input Handling
- [ ] **Dialog UI uses custom debounce with global mutable state** - Location: `cmd/client/dialog_ui.go:186-193` - Impact: `lastKeyState` is a package-level mutable map, not synchronized. While Ebiten calls Update() single-threaded, this pattern shadows `inpututil.IsKeyJustPressed()` which is the correct Ebiten idiom. The custom implementation loses key events if multiple keys are pressed in the same frame, making dialog option selection unreliable
- [ ] **Character creation name input accepts all Unicode including control chars** - Location: `cmd/client/character_creation.go:215-219` - Impact: `ebiten.AppendInputChars()` returns all typed characters including control characters and emoji. No filtering is applied, so names can contain unprintable characters that render as garbage or break text width calculations in the HUD

### Combat System
- [ ] **OnPlayerDamaged uses hardcoded 0.3 multiplier instead of blockReduction field** - Location: `cmd/client/combat.go:682-684` - Impact: `OnPlayerDamaged()` multiplies damage by `0.3` when blocking, but `blockReduction` is initialized to `0.5` at line 89 and correctly used in `CalculateIncomingDamage()` at line 348. This means the external damage callback path reduces damage by 70% while the internal path reduces by 50%, creating inconsistent blocking behavior
- [ ] **Spell cast uses Q key which conflicts with strafe-left** - Location: `cmd/client/combat.go:417` - Impact: Q key is bound to both `ActionStrafeLeft` (line 1065 in main.go) and spell casting (line 417 in combat.go). When the player presses Q, both strafing and spell casting trigger simultaneously, causing unintended movement during magic attacks

### Menu System
- [ ] **Settings menu displays stale values after changes** - Location: `cmd/client/menu.go:118-127` (approx) - Impact: Settings menu items are built as string labels at initialization time (e.g., `fmt.Sprintf("Music Volume: %d", m.volumeLevel)`). When the user changes a setting, the label string is not rebuilt, so the display shows the old value while the underlying setting has changed. User sees wrong values until menu is reopened

### Interaction System
- [ ] **Item/Container detection logic is inverted for full inventories** - Location: `cmd/client/interaction.go:182-187` - Impact: `getInventoryInteractionType()` returns `InteractionItem` when `len(Items) > 0 && Capacity <= len(Items)` (full container), and `InteractionContainer` when `Capacity > len(Items)` (has room). A full container should still be openable as a Container for viewing, but is instead treated as a pickup Item. Entities with full inventories cannot be browsed

### NPC Rendering
- [ ] **NPC sprite cache miss causes silent disappearance** - Location: `cmd/client/npc_render.go:61-64` - Impact: If `GetOrGenerate()` returns nil (sprite generation failure), the NPC is silently skipped with `continue`. No fallback sprite is used and no error is logged. NPCs can vanish from the world without any player-facing indication or debug output

### World/Chunk System
- [ ] **LODManager.SetViewpoint called before lodManager is nil-checked** - Location: `cmd/client/main.go:891` - Impact: `rebuildWorldMap()` calls `g.lodManager.SetViewpoint()` without a nil check on `lodManager`. While `lodManager` is always initialized in `main()`, a future refactor or conditional init could cause a nil panic during chunk map building

### Audio
- [ ] **Ambient audio loops only 2 seconds then stops** - Location: `cmd/client/init.go:193-198` - Impact: `startAmbientAudio()` generates and queues only 2 seconds of audio samples. Once played, the audio stops with no looping mechanism. The game world becomes silent after 2 seconds unless the audio player has internal looping (it doesn't — `Play()` plays the queued buffer once)

## Medium Priority (Polish/Optimization)

### UI Polish
- [ ] **Housing UI propertyMarket nil risk if Open() called before Initialize()** - Location: `cmd/client/housing_ui.go:70-71` - Impact: `propertyMarket` is only set in `Initialize()` but `refreshListings()` accesses it unconditionally. If `Open()` is called before `Initialize()`, nil dereference occurs. Current code always calls Initialize() after NewHousingUI(), but no defensive check exists
- [ ] **PvP UI alpha calculation redundant checks** - Location: `cmd/client/pvp_ui.go:245-256` - Impact: Alpha value calculation has check order `< 50`, `> 255`, `< 0` but the `< 50` clamp means `< 0` is unreachable. While not a crash, the minimum alpha of 50 means notifications never fully fade out, leaving ghost text on screen
- [ ] **Inventory UI slot image reallocated per-frame if slot size changes** - Location: `cmd/client/inventory_ui.go:492-493` - Impact: `drawSlot()` checks if slotImage bounds match `ui.slotSize` and reallocates if not. Since `slotSize` is constant (48), this doesn't trigger in practice, but the `Fill()` call on every slot every frame is wasteful — could be optimized with a pre-rendered slot
- [ ] **Compass direction calculation can produce negative modular index** - Location: `cmd/client/hud.go:429` - Impact: `int((angle + π/8) / (π/4)) % 8` can produce a negative result in Go if the angle normalization fails for edge cases near 0. This would index the `directions` array with a negative index, causing a panic. The angle normalization loop above should prevent this, but the pattern is fragile

### ECS/Systems
- [ ] **Event-driven systems have empty Update() — no way to trigger events** - Location: `pkg/engine/systems/crime.go:853` (BriberySystem), `pkg/engine/systems/skill_progression.go:520` (NPCTrainingSystem), `pkg/engine/systems/vehicle_customization.go` (VehicleCustomizationSystem), `pkg/engine/systems/barrier_collision.go:302` (BarrierDamageSystem) - Impact: These systems have empty `Update()` bodies with comments saying they're "event-driven," but their event trigger methods (e.g., `OnBriberyAccepted()`, `GrantTraining()`) are never called from any client or server code path. The features (bribery, NPC training, vehicle customization, barrier damage) are completely inert
- [ ] **CombatManager creates private system instances not shared with ECS** - Location: `cmd/client/combat.go:75-78` - Impact: `NewCombatManager()` creates `NewCombatSystem()`, `NewProjectileSystem()`, `NewMagicSystem()`, `NewStealthSystem()` instances that are used locally but not registered with the ECS world. Meanwhile, `registerSinglePlayerSystems()` in `init.go:138-142` registers separate instances. Combat state updates happen in two isolated system instances that don't see each other's entity modifications

### Rendering Details
- [ ] **Post-processing writes to framebuffer after WritePixels already uploaded it** - Location: `cmd/client/main.go:1566-1589` - Impact: `Draw()` calls `renderer.Draw(screen)` which writes framebuffer to screen at `draw.go:20`. Then `applyPostProcessing()` reads the framebuffer, modifies it, and writes it back to screen via `WritePixels` again. This double-upload is wasteful and means the first `WritePixels` at line 20 is immediately overwritten. Post-processing should be applied to the framebuffer before the first upload
- [ ] **Particle rendering also double-writes framebuffer** - Location: `cmd/client/main.go:1594-1611` - Impact: Same pattern as post-processing — `drawParticles()` reads the framebuffer, renders particles into it, then uploads again. If both post-processing and particles are active, the framebuffer is uploaded 3 times per frame instead of once

### Configuration
- [ ] **Mouse smoothing with factor=0 disables all mouse input** - Location: `cmd/client/main.go:1243-1244` - Impact: When `SmoothingFactor` is 0 (and `SmoothingOn` is true), the smoothed delta formula becomes `smoothedDeltaX = smoothedDeltaX*1 + deltaX*0 = smoothedDeltaX`. The dead-zone check at line 1246 then zeroes it out. Result: mouse look completely stops responding. Should clamp factor minimum to a small positive value like 0.1

## Resolution Notes

### Dependencies Between Fixes
1. **Renderer blank world** is cosmetic but confusing — consider adding a loading indicator before real chunk data arrives
2. **SetWorldMapDirect stale cells** should be fixed to always keep WorldMap and WorldMapCells in sync
3. **Dialog vocabulary bounds** is future-proofing — verify all 5 genre vocabularies in `pkg/dialog/` have non-empty arrays (currently they do) and add `len()` guards for safety
4. **Combat system duplication** (CombatManager private instances vs ECS-registered instances) is a design issue that affects all combat — fixing this properly requires deciding on single ownership
5. **Post-processing double-upload** and **particle double-upload** should be fixed together by restructuring the draw pipeline to: ClearFramebuffer → drawSky → drawFloorCeiling → drawWalls → applyPostProcessing → drawParticles → single WritePixels → drawSprites → drawHUD
6. **Audio 2-second silence** needs a proper audio loop or continuous generation system

### Build/Test Status
- Both `cmd/client` and `cmd/server` compile successfully
- All 31 test packages pass with `-tags=noebiten`
- No compilation errors or test failures exist — all issues are runtime/logic bugs

### Architecture Notes
- The codebase is well-structured with clean separation between UI, ECS, rendering, and networking
- Most issues are boundary conditions (nil checks, bounds checks) rather than fundamental design flaws
- The dual CombatSystem instance problem (local in CombatManager + registered in ECS) is the most architecturally significant issue
- The stale WorldMapCells issue in SetWorldMapDirect is a data-sync bug that could cause subtle rendering glitches
