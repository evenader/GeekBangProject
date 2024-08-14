package web

import (
	"encoding/json"
	"gitee.com/geekbang/basic-go/webook/internal/domain"
	"gitee.com/geekbang/basic-go/webook/internal/service"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

const (
	emailRegexPattern = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
	// 和上面比起来，用 ` 看起来就比较清爽
	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
)

type UserHandler struct {
	emailRexExp    *regexp.Regexp
	passwordRexExp *regexp.Regexp
	svc            *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{
		emailRexExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRexExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
		svc:            svc,
	}
}

func (h *UserHandler) RegisterRoutes(server *gin.Engine) {
	// REST 风格
	//server.POST("/user", h.SignUp)
	//server.PUT("/user", h.SignUp)
	//server.GET("/users/:username", h.Profile)
	ug := server.Group("/users")
	// POST /users/signup
	ug.POST("/signup", h.SignUp)
	// POST /users/login
	ug.POST("/login", h.Login)
	// POST /users/edit
	ug.POST("/edit", h.Edit)
	// GET /users/profile
	ug.GET("/profile", h.Profile)
}

func (h *UserHandler) SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}

	var req SignUpReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	isEmail, err := h.emailRexExp.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !isEmail {
		ctx.String(http.StatusOK, "非法邮箱格式")
		return
	}

	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusOK, "两次输入密码不对")
		return
	}

	isPassword, err := h.passwordRexExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !isPassword {
		ctx.String(http.StatusOK, "密码必须包含字母、数字、特殊字符，并且不少于八位")
		return
	}

	err = h.svc.Signup(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	switch err {
	case nil:
		ctx.String(http.StatusOK, "注册成功")
	case service.ErrDuplicateEmail:
		ctx.String(http.StatusOK, "邮箱冲突，请换一个")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
}

func (h *UserHandler) Login(ctx *gin.Context) {
	type Req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	u, err := h.svc.Login(ctx, req.Email, req.Password)
	switch err {
	case nil:
		sess := sessions.Default(ctx)
		sess.Set("userId", u.Id)
		sess.Options(sessions.Options{
			// 十五分钟
			MaxAge: 900,
		})
		err = sess.Save()
		if err != nil {
			ctx.String(http.StatusOK, "系统错误")
			return
		}
		ctx.String(http.StatusOK, "登录成功")
	case service.ErrInvalidUserOrPassword:
		ctx.String(http.StatusOK, "用户名或者密码不对")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
}

const nickNameMaxLength = 10
const profileMaxLength = 140

// 第二次作业
func (h *UserHandler) Edit(ctx *gin.Context) {
	type UserInfoEditReq struct {
		Id       int64  `json:"int64"`
		NickName string `json:"nickname"`
		Birthday string `json:"birthday"`
		Profile  string `json:"profile"`
	}

	var req UserInfoEditReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	if len(req.NickName) > nickNameMaxLength {
		ctx.String(http.StatusOK, "昵称长度超过%d，请重新输入", nickNameMaxLength)
		return
	}

	t, err := time.Parse("2006-01-02", req.Birthday)
	if err != nil {
		ctx.String(http.StatusOK, "生日格式错误，请重新输入")
		return
	}

	if len(req.NickName) > profileMaxLength {
		ctx.String(http.StatusOK, "简介长度超过%d，请重新输入！", profileMaxLength)
		return
	}

	u := &domain.User{
		NickName: req.NickName,
		Birthday: t.Unix(),
		Profile:  req.Profile,
		Id:       req.Id,
	}

	err = h.svc.Edit(ctx, u)
	switch err {
	case nil:
		ctx.String(http.StatusOK, "编辑成功")
	case service.ErrInvalidId:
		ctx.String(http.StatusOK, "id错误")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
}

func (h *UserHandler) Profile(ctx *gin.Context) {
	originId := ctx.Query("id")

	id, err := strconv.Atoi(originId)
	if err != nil {
		ctx.String(http.StatusOK, "id格式错误")
		return
	}

	u, err := h.svc.Profile(ctx, int64(id))
	switch err {
	case nil:
		jsonBytes, err := json.Marshal(u)
		if err != nil {
			ctx.String(http.StatusOK, "json序列化失败")
		}
		jsonString := string(jsonBytes)
		ctx.String(http.StatusOK, jsonString)
	case service.ErrInvalidId:
		ctx.String(http.StatusOK, "id错误")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
}
