package middlewares

import (
	"encoding/json"
	"strconv"
	"strings"

	"go-mvc/dto"
	"go-mvc/services"

	beegoCtx "github.com/beego/beego/v2/server/web/context"
)

// JWTAuthFilter returns a Beego filter handler that checks JWT Bearer tokens
func JWTAuthFilter(authService services.AuthService) func(ctx *beegoCtx.Context) {
	return func(ctx *beegoCtx.Context) {
		// Allow OPTIONS requests for CORS preflight
		if ctx.Input.Method() == "OPTIONS" {
			return
		}

		authHeader := ctx.Input.Header("Authorization")
		if authHeader == "" || !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			respondUnauthorized(ctx, "Authorization token is missing or malformed")
			return
		}

		tokenString := strings.TrimSpace(authHeader[7:])
		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			respondUnauthorized(ctx, "Invalid or expired token")
			return
		}

		userID, err := strconv.Atoi(claims.Subject)
		if err != nil {
			respondUnauthorized(ctx, "Invalid token subject")
			return
		}

		ctx.Input.SetData("userID", userID)
	}
}

func respondUnauthorized(ctx *beegoCtx.Context, message string) {
	ctx.Output.SetStatus(401)
	ctx.Output.Header("Content-Type", "application/json")
	resp, _ := json.Marshal(dto.APIResponse{
		Success: false,
		Message: message,
	})
	_ = ctx.Output.Body(resp)
}
