package tokenmanager

import (
	"URLShortener/internal/core/models"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type RepositoryMock struct {
	SetTokenCall, GetTokenCall, DeleteTokenCall, GetUserCall, CreateUserCall bool
	HashedPassword, UserName, Email                                          string
}

func (r *RepositoryMock) SetToken(ctx context.Context, key, value string, ttl time.Duration) error {
	r.SetTokenCall = true
	return nil
}
func (r *RepositoryMock) GetToken(ctx context.Context, key string) (string, error) {
	r.GetTokenCall = true

	return "", nil
}
func (r *RepositoryMock) DeleteToken(ctx context.Context, key string) error {
	r.DeleteTokenCall = true
	return nil
}
func (r *RepositoryMock) GetUser(ctx context.Context, email string) (*models.User, error) {
	r.GetUserCall = true

	return &models.User{
		ID:       0,
		Username: r.UserName,
		Email:    r.Email,
		HashPass: r.HashedPassword,
	}, nil
}
func (r *RepositoryMock) CreateUser(ctx context.Context, username, email, hash string) error {
	r.CreateUserCall = true
	r.UserName = username
	r.Email = email
	r.HashedPassword = hash
	return nil
}

func TestCreateNewUser(t *testing.T) {

	cases := []struct {
		Name        string
		Username    string
		Email       string
		Password    string
		ExpectedErr error
	}{{
		Name:        "Normal Test",
		Username:    "Bebrikov",
		Email:       "B3brikov@email",
		Password:    "TestPassword",
		ExpectedErr: nil,
	}, {
		Name:        "Empty password",
		Username:    "Bebrikov",
		Email:       "B3brikov@email",
		Password:    "",
		ExpectedErr: EmptyPassword,
	}}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			repo := &RepositoryMock{}
			manager := NewTokenManager(repo, 4*time.Second, 1*time.Minute, []byte("test"))

			err := manager.CreateNewUser(context.TODO(), c.Username, c.Email, c.Password)
			if c.ExpectedErr != nil {
				assert.ErrorIs(t, err, c.ExpectedErr)
				require.False(t, repo.CreateUserCall)
			} else {
				require.NoError(t, err)
				require.NoError(t, bcrypt.CompareHashAndPassword([]byte(repo.HashedPassword), []byte(c.Password)))
				assert.Equal(t, repo.Email, c.Email)
				assert.Equal(t, repo.UserName, c.Username)
				require.True(t, repo.CreateUserCall)
			}

		})
	}
}

func TestAuthorize(t *testing.T) {

	testCases := []struct {
		desc, Email, Password string
		ExpErr                error
	}{
		{
			desc:     "Empty test",
			Email:    "",
			Password: "",
			ExpErr:   EmptyEmail,
		}, {
			desc:     "Correct val",
			Email:    "correct",
			Password: "correct",
			ExpErr:   nil,
		}, {
			desc:     "Empty pass",
			Email:    "Some email",
			Password: "",
			ExpErr:   EmptyPassword,
		},
	}
	pass, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			repo := &RepositoryMock{
				SetTokenCall:    false,
				GetTokenCall:    false,
				DeleteTokenCall: false,
				GetUserCall:     false,
				CreateUserCall:  false,
				HashedPassword:  string(pass),
				UserName:        "",
				Email:           "correct",
			}
			manager := NewTokenManager(repo, 4*time.Second, 1*time.Minute, []byte("test"))

			tokens, err := manager.Authorize(context.TODO(), tC.Email, tC.Password)
			if tC.ExpErr != nil {
				assert.ErrorIs(t, err, tC.ExpErr)
			} else {
				assert.Equal(t, repo.Email, tC.Email)
				require.NoError(t, bcrypt.CompareHashAndPassword(pass, []byte(tC.Password)))
				require.NotEmpty(t, tokens)
			}
		})
	}
}
