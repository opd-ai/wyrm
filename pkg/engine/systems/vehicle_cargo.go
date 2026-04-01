package systems

import "fmt"

// CargoContainer represents any vehicle or entity that can hold cargo.
type CargoContainer interface {
	GetCargo() map[string]float64
	GetCargoCapacity() float64
}

// calculateCargoTotal returns the total cargo weight in the container.
func calculateCargoTotal(cargo map[string]float64) float64 {
	total := 0.0
	for _, amount := range cargo {
		total += amount
	}
	return total
}

// loadCargoToContainer adds cargo to a container, checking capacity.
func loadCargoToContainer(cargo map[string]float64, cargoCapacity float64, item string, amount float64) error {
	currentCargo := calculateCargoTotal(cargo)
	if currentCargo+amount > cargoCapacity {
		return fmt.Errorf("cargo capacity exceeded")
	}
	cargo[item] += amount
	return nil
}

// unloadCargoFromContainer removes cargo from a container.
func unloadCargoFromContainer(cargo map[string]float64, item string, amount float64) error {
	if cargo[item] < amount {
		return fmt.Errorf("not enough cargo")
	}
	cargo[item] -= amount
	if cargo[item] <= 0 {
		delete(cargo, item)
	}
	return nil
}

// PassengerContainer represents any vehicle that can hold passengers.
type PassengerContainer interface {
	GetPassengers() []any
	GetPassengerCapacity() int
}

// boardPassengerToSlice adds a passenger to a slice, checking capacity and duplicates.
// Returns the updated slice and any error.
func boardPassengerToSlice[T comparable](passengers []T, capacity int, passenger T) ([]T, error) {
	if len(passengers) >= capacity {
		return passengers, fmt.Errorf("passenger capacity full")
	}
	for _, p := range passengers {
		if p == passenger {
			return passengers, fmt.Errorf("already aboard")
		}
	}
	return append(passengers, passenger), nil
}

// disembarkPassengerFromSlice removes a passenger from a slice.
// Returns the updated slice and any error.
func disembarkPassengerFromSlice[T comparable](passengers []T, passenger T) ([]T, error) {
	for i, p := range passengers {
		if p == passenger {
			return append(passengers[:i], passengers[i+1:]...), nil
		}
	}
	return passengers, fmt.Errorf("passenger not found")
}

// refuelVehicle adds fuel to a vehicle with capacity checking.
func refuelVehicle(currentFuel *float64, fuelCapacity, amount float64) error {
	if fuelCapacity == 0 {
		return fmt.Errorf("vehicle does not use fuel")
	}
	*currentFuel += amount
	if *currentFuel > fuelCapacity {
		*currentFuel = fuelCapacity
	}
	return nil
}

// damageVehicleHealth reduces health/hull by damage amount, clamping to 0.
func damageVehicleHealth(currentHealth *int, damage int) {
	*currentHealth -= damage
	if *currentHealth < 0 {
		*currentHealth = 0
	}
}

// repairVehicleHealth increases health/hull by repair amount, clamping to max.
func repairVehicleHealth(currentHealth *int, maxHealth, repair int) {
	*currentHealth += repair
	if *currentHealth > maxHealth {
		*currentHealth = maxHealth
	}
}
