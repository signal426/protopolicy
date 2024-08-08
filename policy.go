package propl

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/protobuf/proto"
)

type Precheck[T proto.Message] func(ctx context.Context, msg T) error

type Subject interface {
	HasTrait(t Trait) bool
	ConditionalAction(condition Condition) Action
}

type Trait interface {
	And() Trait
	Or() Trait
	InfractionsString() string
	Type() TraitType
	Valid() bool
}

type Policy struct {
	conditions Condition
	traits     Trait
}

// Execute checks traits on the field based on the conditional action signal
// returned from the subject.
func (p *Policy) Execute(s Subject) error {
	switch s.ConditionalAction(p.conditions) {
	case Skip:
		return nil
	case Fail:
		return fmt.Errorf("subject did not meet conditions %s", p.conditions.FlagsString())
	default:
		return p.checkTraits(s, p.traits)
	}
}

func (p *Policy) checkTraits(s Subject, t Trait) error {
	if t == nil {
		return nil
	}
	if t.Valid() && !s.HasTrait(t) {
		// if we have an or, keep going
		if t.Or().Valid() {
			return p.checkTraits(s, t.Or())
		}
		// else, we're done checking
		return errors.New(t.InfractionsString())
	}
	// if there's an and condition, keep going
	// else, we're done
	if t.And().Valid() {
		return p.checkTraits(s, t.And())
	}
	return nil
}
