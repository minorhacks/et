package et

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Example namespace
type porridgeErrs struct {
	Namespace
}

type errTooCold struct {
	Inside[porridgeErrs]
}

type errTooHot struct {
	Inside[porridgeErrs]
}

type chairErrs struct {
	Namespace
}

type errTooBig struct {
	Inside[chairErrs]
}

type errTooSmall struct {
	Inside[chairErrs]
}

type bedErrs struct {
	Namespace
}

type errTooHard struct {
	Inside[bedErrs]
}

type errTooSoft struct {
	Inside[bedErrs]
}

func TestErrtag(t *testing.T) {
	errA := Errorf[errTooHot]("ouch my tongue")
	assert.Equal(t, "porridgeErrs::errTooHot: ouch my tongue", errA.Error())

	errB := Errorf[errTooSmall]("cannot fit")
	assert.Equal(t, "chairErrs::errTooSmall: cannot fit", errB.Error())

	errC := Errorf[errTooSoft]("floof")
	assert.Equal(t, "bedErrs::errTooSoft: floof", errC.Error())

	papaBearIssues := Wrap[errTooHard](Wrap[errTooBig](Wrap[errTooHot](errors.New("not for me"))))
	assert.Equal(t, "bedErrs::errTooHard: chairErrs::errTooBig: porridgeErrs::errTooHot: not for me", papaBearIssues.Error())

	assert.ErrorIs(t, papaBearIssues, OfType[errTooHard]())
	assert.ErrorIs(t, papaBearIssues, OfType[errTooBig]())
	assert.ErrorIs(t, papaBearIssues, OfType[errTooHot]())

	assert.ErrorIs(t, papaBearIssues, OfKind[porridgeErrs]())
	assert.ErrorIs(t, papaBearIssues, OfKind[chairErrs]())
	assert.ErrorIs(t, papaBearIssues, OfKind[bedErrs]())

	porridgeErr := AsKind[porridgeErrs]()
	assert.ErrorAs(t, papaBearIssues, &porridgeErr)
	assert.ErrorIs(t, porridgeErr, OfType[errTooHot]())
}

