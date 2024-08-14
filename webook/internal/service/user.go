package service

import (
	"context"
	"errors"
	"gitee.com/geekbang/basic-go/webook/internal/domain"
	"gitee.com/geekbang/basic-go/webook/internal/repository"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail        = repository.ErrDuplicateEmail
	ErrInvalidUserOrPassword = errors.New("用户不存在或者密码不对")
	ErrInvalidId             = errors.New("用户id错误")
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (svc *UserService) Signup(ctx context.Context, u domain.User) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return svc.repo.Create(ctx, u)
}

func (svc *UserService) Login(ctx context.Context, email string, password string) (domain.User, error) {
	u, err := svc.repo.FindByEmail(ctx, email)
	if err == repository.ErrUserNotFound {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	if err != nil {
		return domain.User{}, err
	}
	// 检查密码对不对
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return u, nil
}

func (svc *UserService) Edit(ctx context.Context, u *domain.User) error {
	originUser, err := svc.repo.FindById(ctx, u.Id)
	if err == repository.ErrRecordNotFound {
		return ErrInvalidId
	}
	if err != nil {
		return err
	}

	if u.NickName != "" {
		originUser.NickName = u.NickName
	}
	if u.Birthday != 0 {
		originUser.Birthday = u.Birthday
	}
	if u.Profile != "" {
		originUser.Profile = u.Profile
	}

	err = svc.repo.UpdateById(ctx, originUser)
	if err != nil {
		return err
	}

	return nil
}

func (svc *UserService) Profile(ctx *gin.Context, id int64) (domain.User, error) {
	u, err := svc.repo.FindById(ctx, id)
	if err == repository.ErrRecordNotFound {
		return domain.User{}, ErrInvalidId
	}
	if err != nil {
		return domain.User{}, err
	}
	return u, nil
}
