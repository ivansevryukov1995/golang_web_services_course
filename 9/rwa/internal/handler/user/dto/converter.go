package dto

import (
	"rwa/internal/model"
	userService "rwa/internal/service/user"
)

func ToUserRegResponse(u model.User, token string) UserResponse {
	return UserResponse{
		User: User{
			Email:     u.Email,
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
			Username:  u.Username,
			Token:     token,
		},
	}
}

func ToUserLoginResponse(u model.User, token string) UserResponse {
	return UserResponse{
		User: User{
			Email:     u.Email,
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
			Username:  u.Username,
			Token:     token,
		},
	}
}

func ToUserCurrentResponse(u model.User, token string) UserResponse {
	return UserResponse{
		User: User{
			Email:     u.Email,
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
			Username:  u.Username,
			Bio:       u.Bio,
			Token:     token,
		},
	}
}

func ToUserUpdateResponse(u model.User, token string) UserResponse {
	return UserResponse{
		User: User{
			Email:     u.Email,
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
			Username:  u.Username,
			Bio:       u.Bio,
			Token:     token,
		},
	}
}

func ToUserInput(req UserRequest) userService.UserInput {
	return userService.UserInput{
		Email:    req.User.Email,
		Password: req.User.Password,
		Username: req.User.Username,
		Bio:      req.User.Bio,
	}
}
