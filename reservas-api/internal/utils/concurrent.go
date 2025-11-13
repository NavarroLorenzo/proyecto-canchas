package utils

import (
	"fmt"
	"reservas-api/internal/dto"
	"sync"
	"time"
)

// ConcurrentValidation representa una validación a ejecutar concurrentemente
type ConcurrentValidation struct {
	Name     string
	Function func() dto.ValidationResult
}

// ExecuteConcurrentValidations ejecuta múltiples validaciones en paralelo
// usando GoRoutines, Channels y WaitGroup
func ExecuteConcurrentValidations(validations []ConcurrentValidation) (bool, []string) {
	// Canal para recibir resultados
	resultsChan := make(chan dto.ValidationResult, len(validations))

	// WaitGroup para sincronizar las goroutines
	var wg sync.WaitGroup

	// Lanzar una goroutine por cada validación
	for _, validation := range validations {
		wg.Add(1)

		// Ejecutar validación en goroutine
		go func(v ConcurrentValidation) {
			defer wg.Done()

			// Ejecutar la función de validación
			result := v.Function()

			// Enviar resultado al canal
			resultsChan <- result
		}(validation)
	}

	// Cerrar el canal cuando todas las goroutines terminen
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Recolectar resultados del canal
	allValid := true
	var errors []string

	for result := range resultsChan {
		if !result.Valid {
			allValid = false
			errors = append(errors, result.Message)
		}
	}

	return allValid, errors
}

// CalculateDuration calcula la duración en minutos entre dos horas
func CalculateDuration(startTime, endTime string) (int, error) {
	layout := "15:04"

	start, err := time.Parse(layout, startTime)
	if err != nil {
		return 0, fmt.Errorf("invalid start time format: %w", err)
	}

	end, err := time.Parse(layout, endTime)
	if err != nil {
		return 0, fmt.Errorf("invalid end time format: %w", err)
	}

	if end.Before(start) || end.Equal(start) {
		return 0, fmt.Errorf("end time must be after start time")
	}

	duration := end.Sub(start)
	return int(duration.Minutes()), nil
}

// CalculatePrice devuelve el precio total del turno.
// El valor almacenado en la cancha ya representa el costo completo.
func CalculatePrice(pricePerTurn float64, _ int) float64 {
	return pricePerTurn
}

// ParseDate convierte un string de fecha a time.Time
func ParseDate(dateStr string) (time.Time, error) {
	layout := "2006-01-02"
	return time.Parse(layout, dateStr)
}

// FormatDate convierte un time.Time a string
func FormatDate(date time.Time) string {
	return date.Format("2006-01-02")
}
