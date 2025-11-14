package utils

import "strings"

// NormalizeString convierte a minúsculas y reemplaza caracteres acentuados
// por su versión sin acento para evitar problemas de matching.
func NormalizeString(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)

	replacer := strings.NewReplacer(
		"á", "a",
		"é", "e",
		"í", "i",
		"ó", "o",
		"ú", "u",
		"ü", "u",
		"Á", "a",
		"É", "e",
		"Í", "i",
		"Ó", "o",
		"Ú", "u",
		"Ü", "u",
		"ñ", "n",
		"Ñ", "n",
	)

	return replacer.Replace(s)
}
