package domains

import "testing"

// El PK-ID lleva el COM adentro para que el identificador se explique solo:
// quien lo ve en el ERP o en engram sabe de qué oportunidad es sin cruzar
// ninguna tabla. Va SIN guion porque en el PK-ID el guion separa campos.
func TestElPKIDLlevaElCOMAdentro(t *testing.T) {
	if got := composePKID("PK", "PROP", 2026, 14, "COM-989"); got != "PK-PROP-2026-0014-COM989" {
		t.Errorf("composePKID = %q", got)
	}
}

func TestSinCOMElPKIDQuedaComoSiempre(t *testing.T) {
	// Los tipos que no cuelgan de una oportunidad (lesson, session) no lo llevan.
	if got := composePKID("PK", "LES", 2026, 7, ""); got != "PK-LES-2026-0007" {
		t.Errorf("composePKID = %q", got)
	}
}

func TestElCOMSeNormalizaAntesDeEntrar(t *testing.T) {
	for _, crudo := range []string{"com-989", "  COM-989  ", "COM989", "com 989"} {
		if got := composePKID("PK", "PROP", 2026, 14, crudo); got != "PK-PROP-2026-0014-COM989" {
			t.Errorf("composePKID(%q) = %q", crudo, got)
		}
	}
}

// Una clave que no es de la cadena comercial no se cuela: preferimos un PK-ID
// sin sufijo antes que uno que mienta sobre de qué oportunidad viene.
func TestUnaClaveQueNoEsCOMNoEntra(t *testing.T) {
	for _, basura := range []string{"AAC-123", "GDC-55", "banco de condensadores", "COM-", "COM-0"} {
		if got := composePKID("PK", "PROP", 2026, 14, basura); got != "PK-PROP-2026-0014" {
			t.Errorf("composePKID(%q) = %q, no debería llevar sufijo", basura, got)
		}
	}
}

// El cálculo anterior contaba las filas del tipo SIN filtrar por año: la serie
// de 2027 continuaba la de 2026 para siempre. Y contar en vez de mirar el
// máximo devuelve un id YA EMITIDO apenas alguien borra una referencia.
func TestElSecuencialMiraElMaximoDelAnio(t *testing.T) {
	existentes := []string{
		"PK-PROP-2026-0001-COM1",
		"PK-PROP-2026-0009-COM9", // falta el medio: contar daría 3
		"PK-PROP-2025-0400",      // otro año
		"PK-PROJ-2026-0088",      // otro tipo
		"basura",
		"",
	}
	if got := nextSequential(existentes, "PK", "PROP", 2026); got != 10 {
		t.Errorf("nextSequential = %d, want 10", got)
	}
}

func TestElPrimeroDelAnioEsElUno(t *testing.T) {
	if got := nextSequential(nil, "PK", "PROP", 2026); got != 1 {
		t.Errorf("nextSequential vacío = %d, want 1", got)
	}
	if got := nextSequential([]string{"PK-PROP-2026-0571-COM940"}, "PK", "PROP", 2027); got != 1 {
		t.Errorf("nextSequential en año nuevo = %d, want 1", got)
	}
}

// Los PK-ID viejos sin COM conviven con los nuevos y el secuencial los cuenta.
func TestConviveConLosPKIDViejos(t *testing.T) {
	if got := nextSequential([]string{"PK-PROP-2026-0293"}, "PK", "PROP", 2026); got != 294 {
		t.Errorf("nextSequential = %d, want 294", got)
	}
}
