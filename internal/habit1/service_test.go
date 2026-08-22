package habit1

import (
	"context"
	"testing"
	"time"
)

// testUserID es el UID de Firebase que el handler obtendría del token.
const testUserID = "firebase-uid-123"

// fakeRepository es un doble de prueba en memoria para Repository.
// created guarda los registros en el orden en que se persistieron, para poder
// afirmar que el del día siguiente se crea después del que envió el usuario.
type fakeRepository struct {
	created    []*Habit1
	err        error
	listedUser string
}

func (r *fakeRepository) Create(_ context.Context, h *Habit1) (*Habit1, error) {
	if r.err != nil {
		return nil, r.err
	}
	h.ID = "generated-id"
	r.created = append(r.created, h)
	return h, nil
}

func (r *fakeRepository) ListByUser(_ context.Context, userID string) ([]*Habit1, error) {
	r.listedUser = userID
	if r.err != nil {
		return nil, r.err
	}
	out := make([]*Habit1, 0, len(r.created))
	for _, h := range r.created {
		if h.UserID == userID {
			out = append(out, h)
		}
	}
	return out, nil
}

// register registra el ejemplo de la especificación: repetición 5 del
// 2026-08-22, despertando a las 06:00.
func register(t *testing.T, repo *fakeRepository) *Habit1 {
	t.Helper()

	fecha := time.Date(2026, time.August, 22, 0, 0, 0, 0, time.UTC)
	record, err := NewService(repo).RegisterHabit1(
		context.Background(), testUserID, 5, fecha, "06:00", "22:30", "07:30", true, true,
	)
	if err != nil {
		t.Fatalf("RegisterHabit1() error = %v, want nil", err)
	}
	return record
}

func TestRegisterHabit1PersistsTheRecordSentByTheUser(t *testing.T) {
	repo := &fakeRepository{}

	record := register(t, repo)

	if len(repo.created) != 2 {
		t.Fatalf("created records = %d, want 2 (el del usuario y el del día siguiente)", len(repo.created))
	}

	got := repo.created[0]
	if got != record {
		t.Errorf("RegisterHabit1() devolvió %+v, want el registro del usuario %+v", record, got)
	}
	if got.UserID != testUserID {
		t.Errorf("UserID = %q, want %q", got.UserID, testUserID)
	}
	if got.NumeroDeRepeticion != 5 {
		t.Errorf("NumeroDeRepeticion = %d, want 5", got.NumeroDeRepeticion)
	}
	if got.HoraDespertar != "06:00" || got.HoraDormir != "22:30" || got.HorasDormidas != "07:30" {
		t.Errorf("horas = %q/%q/%q, want 06:00/22:30/07:30", got.HoraDespertar, got.HoraDormir, got.HorasDormidas)
	}
	if !got.RitualNoche || !got.RitualDia {
		t.Errorf("rituales = %v/%v, want true/true", got.RitualNoche, got.RitualDia)
	}
}

func TestRegisterHabit1CreatesTheNextDayRecord(t *testing.T) {
	repo := &fakeRepository{}

	register(t, repo)

	if len(repo.created) != 2 {
		t.Fatalf("created records = %d, want 2", len(repo.created))
	}

	next := repo.created[1]
	wantFecha := time.Date(2026, time.August, 23, 0, 0, 0, 0, time.UTC)

	if next.UserID != testUserID {
		t.Errorf("UserID = %q, want %q", next.UserID, testUserID)
	}
	if next.NumeroDeRepeticion != 6 {
		t.Errorf("NumeroDeRepeticion = %d, want 6", next.NumeroDeRepeticion)
	}
	if !next.Fecha.Equal(wantFecha) {
		t.Errorf("Fecha = %v, want %v", next.Fecha, wantFecha)
	}
	if next.HoraDespertar != "05:45" {
		t.Errorf("HoraDespertar = %q, want 05:45 (15 minutos antes)", next.HoraDespertar)
	}
	if next.HoraDormir != ZeroTimeOfDay {
		t.Errorf("HoraDormir = %q, want %q", next.HoraDormir, ZeroTimeOfDay)
	}
	if next.HorasDormidas != ZeroTimeOfDay {
		t.Errorf("HorasDormidas = %q, want %q", next.HorasDormidas, ZeroTimeOfDay)
	}
	if next.RitualNoche || next.RitualDia {
		t.Errorf("rituales = %v/%v, want false/false", next.RitualNoche, next.RitualDia)
	}
}

// La hora de despertar es una hora del reloj: restar 15 minutos a 00:05 da la
// vuelta al día en vez de salirse del rango.
func TestRegisterHabit1WrapsTheWakeUpTimeAroundMidnight(t *testing.T) {
	repo := &fakeRepository{}
	fecha := time.Date(2026, time.August, 22, 0, 0, 0, 0, time.UTC)

	if _, err := NewService(repo).RegisterHabit1(
		context.Background(), testUserID, 1, fecha, "00:05", "22:00", "07:00", false, false,
	); err != nil {
		t.Fatalf("RegisterHabit1() error = %v, want nil", err)
	}

	if got := repo.created[1].HoraDespertar; got != "23:50" {
		t.Errorf("HoraDespertar = %q, want 23:50", got)
	}
}

func TestRegisterHabit1RejectsInvalidRecords(t *testing.T) {
	fecha := time.Date(2026, time.August, 22, 0, 0, 0, 0, time.UTC)

	cases := []struct {
		name               string
		userID             string
		numeroDeRepeticion int
		fecha              time.Time
		horaDespertar      string
		horaDormir         string
		horasDormidas      string
		want               error
	}{
		{"sin dueño", "", 1, fecha, "06:00", "22:00", "07:00", ErrInvalidUserID},
		{"repetición cero", testUserID, 0, fecha, "06:00", "22:00", "07:00", ErrInvalidNumeroDeRepeticion},
		{"sin fecha", testUserID, 1, time.Time{}, "06:00", "22:00", "07:00", ErrInvalidFecha},
		{"hora de despertar vacía", testUserID, 1, fecha, "", "22:00", "07:00", ErrInvalidHoraDespertar},
		{"hora de despertar fuera de rango", testUserID, 1, fecha, "24:00", "22:00", "07:00", ErrInvalidHoraDespertar},
		{"hora de dormir sin minutos", testUserID, 1, fecha, "06:00", "22", "07:00", ErrInvalidHoraDormir},
		{"horas dormidas con minutos inválidos", testUserID, 1, fecha, "06:00", "22:00", "07:60", ErrInvalidHorasDormidas},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			repo := &fakeRepository{}

			_, err := NewService(repo).RegisterHabit1(
				context.Background(), c.userID, c.numeroDeRepeticion, c.fecha,
				c.horaDespertar, c.horaDormir, c.horasDormidas, false, false,
			)

			if err != c.want {
				t.Errorf("error = %v, want %v", err, c.want)
			}
			if len(repo.created) != 0 {
				t.Errorf("created records = %d, want 0: un registro inválido no persiste nada", len(repo.created))
			}
		})
	}
}

func TestGetHabit1RecordsOnlyQueriesTheAuthenticatedUser(t *testing.T) {
	repo := &fakeRepository{}
	register(t, repo)

	records, err := NewService(repo).GetHabit1Records(context.Background(), testUserID)
	if err != nil {
		t.Fatalf("GetHabit1Records() error = %v, want nil", err)
	}

	if repo.listedUser != testUserID {
		t.Errorf("listedUser = %q, want %q", repo.listedUser, testUserID)
	}
	if len(records) != 2 {
		t.Errorf("records = %d, want 2", len(records))
	}
}

func TestGetHabit1RecordsWithoutOwnerDoesNotQuery(t *testing.T) {
	repo := &fakeRepository{}

	_, err := NewService(repo).GetHabit1Records(context.Background(), "")

	if err != ErrInvalidUserID {
		t.Errorf("error = %v, want %v", err, ErrInvalidUserID)
	}
	if repo.listedUser != "" {
		t.Errorf("listedUser = %q, want vacío: sin dueño no se consulta nada", repo.listedUser)
	}
}
