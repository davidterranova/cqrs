package cqrs_test

import (
	"fmt"
	"testing"

	"github.com/davidterranova/cqrs/internal/cqrs"
	"github.com/davidterranova/cqrs/internal/cqrs/memory"
	"github.com/stretchr/testify/assert"
)

type user struct {
	*cqrs.BaseAggregate
	username string
}

type cmdCreateUser struct {
	*cqrs.BaseCommand

	Username string
}

type cmdUpdateUser struct {
	*cqrs.BaseCommand

	Username string
}

type evtUserCreated struct {
	username string
}

type evtUserUpdatedUsername struct {
	username string
}

func (c cmdCreateUser) Handle(a cqrs.Aggregate) ([]cqrs.Event, error) {
	if a.Version() != 0 {
		return nil, cqrs.ErrInvalidAggregateVersion
	}

	return []cqrs.Event{
		cqrs.NewEvent().WithAggregate(a).WithEventData(&evtUserCreated{username: c.Username}),
	}, nil
}

func (c cmdUpdateUser) Handle(a cqrs.Aggregate) ([]cqrs.Event, error) {
	if a.Version() == 0 {
		return nil, cqrs.ErrInvalidAggregateVersion
	}

	return []cqrs.Event{
		cqrs.NewEvent().WithAggregate(a).WithEventData(&evtUserUpdatedUsername{username: c.Username}),
	}, nil
}

func (d evtUserCreated) Route(e cqrs.Event, a cqrs.Aggregate) {
	u, ok := a.(*user)
	if !ok {
		panic(fmt.Errorf("%w: %T", cqrs.ErrInvalidAggregateType, a))
	}

	u.username = d.username
}

func (d evtUserUpdatedUsername) Route(e cqrs.Event, a cqrs.Aggregate) {
	u, ok := a.(*user)
	if !ok {
		panic(fmt.Errorf("%w: %T", cqrs.ErrInvalidAggregateType, a))
	}

	u.username = d.username
}

func TestAggregateHandler(t *testing.T) {
	userHandler := cqrs.NewAggregateHandler(
		memory.NewEventStore(),
		memory.NewEventRegistry(),
		func() cqrs.Aggregate { return &user{BaseAggregate: cqrs.NewBaseAggregate()} },
	)

	// create user
	createUser := &cmdCreateUser{
		BaseCommand: cqrs.NewBaseCommand(nil),
		Username:    "dterranova",
	}

	t.Logf("command:\t %+v\n", createUser)

	auser, err := userHandler.HandleCommand(createUser)
	assert.NoError(t, err)
	assert.NotNil(t, auser)

	t.Logf("aggregate:\t %+v\n", auser)

	assert.Equal(t, auser.Version(), 1)

	u, ok := auser.(*user)
	assert.True(t, ok)
	assert.Equal(t, createUser.Username, u.username)

	// update user
	updateUser := &cmdUpdateUser{
		BaseCommand: cqrs.NewBaseCommand(auser),
		Username:    "davidterranova",
	}
	auser2, err := userHandler.HandleCommand(updateUser)
	assert.NoError(t, err)
	assert.NotNil(t, auser2)

	t.Logf("aggregate:\t %+v\n", auser2)

	assert.Equal(t, auser2.Version(), 2)

	u, ok = auser2.(*user)
	assert.True(t, ok)
	assert.Equal(t, updateUser.Username, u.username)
	assert.Equal(t, auser.ID(), auser2.ID())
	assert.Equal(t, auser.CreatedAt(), auser2.CreatedAt())
}
