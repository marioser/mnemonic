package domains

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Sólo COM identifica una oportunidad comercial. El Jira de S&G tiene más
// tableros —AAC de soporte, GDC de compras, GDP y GDI de las etapas
// siguientes— y ninguno nombra el origen de una oferta. Tolerante con cómo la
// escribe una persona ("com 989", "COM989"), estricta con qué acepta.
var comKey = regexp.MustCompile(`^(?i:COM)[\s_-]*(\d+)$`)

// normalizeCOM devuelve la forma compacta que entra en el PK-ID ("COM989"), o
// "" si lo recibido no es una clave de oportunidad.
//
// NO CORRIGE POR PARECIDO: quien escribió "AAC-123" nombró un issue que existe,
// sólo que de otro tablero. Meterlo igual produce un identificador que miente
// sobre de qué oportunidad viene, y eso no se detecta nunca.
func normalizeCOM(raw string) string {
	m := comKey.FindStringSubmatch(strings.TrimSpace(raw))
	if m == nil {
		return ""
	}
	n, err := strconv.Atoi(m[1])
	if err != nil || n < 1 { // Jira numera desde 1: un 0 es un typo
		return ""
	}
	return fmt.Sprintf("COM%d", n)
}

// composePKID arma PK-{TIPO}-{AÑO}-{NNNN}[-COM{N}].
//
// El COM va adentro para que el identificador se explique solo: quien ve un
// ref_ext en el ERP o un topic_key en engram sabe de qué oportunidad es sin
// cruzar ninguna tabla. Sin guion propio, porque en el PK-ID el guion separa
// campos y dos guiones con significados distintos obligan a parsear contando
// desde el final.
func composePKID(prefix, refType string, year, seq int, com string) string {
	id := fmt.Sprintf("%s-%s-%d-%04d", prefix, refType, year, seq)
	if c := normalizeCOM(com); c != "" {
		return id + "-" + c
	}
	return id
}

var pkIDSeq = regexp.MustCompile(`^([A-Z]+)-([A-Z]+)-(\d{4})-(\d+)`)

// nextSequential devuelve el próximo número de la serie de ESTE tipo y ESTE año.
//
// Dos cosas que el cálculo anterior hacía mal, y las dos emiten identificadores
// duplicados —lo peor que puede hacer un identificador—:
//
//   - Contaba las filas del tipo SIN filtrar por año, así que la serie no
//     reiniciaba en enero y la de 2027 continuaba la de 2026.
//   - Contar en vez de mirar el máximo devuelve un id YA EMITIDO apenas alguien
//     borra una referencia del medio, y el upsert lo pisa sin avisar.
func nextSequential(existing []string, prefix, refType string, year int) int {
	max := 0
	for _, id := range existing {
		m := pkIDSeq.FindStringSubmatch(id)
		if m == nil || m[1] != prefix || m[2] != refType {
			continue
		}
		if y, err := strconv.Atoi(m[3]); err != nil || y != year {
			continue
		}
		if n, err := strconv.Atoi(m[4]); err == nil && n > max {
			max = n
		}
	}
	return max + 1
}
