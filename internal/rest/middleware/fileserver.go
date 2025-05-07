package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest"
	auth "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
	gen "github.com/go-park-mail-ru/2025_1_SuperChips/protos/gen/auth"
)

func Fileserver(ctx context.Context, UserService gen.AuthClient) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			claims, _ := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)

			filePath := strings.Split(r.URL.Path, "/")
			if len(filePath) < 2 {
				log.Println("less than 2")
				rest.HttpErrorToJson(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}

			if filePath[0] != "img" {	
				log.Println("not img")
				next.ServeHTTP(w, r)
			} else {
				log.Println("checking permission")
				resp, err := UserService.CheckImgPermission(ctx, &gen.CheckImgPermissionRequest{
					ImageName: filePath[1],
					ID: int64(claims.UserID),
				})
				if err != nil {
					log.Printf("permission err %v", err)
					rest.HttpErrorToJson(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
					return
				}

				hasAccess := resp.HasAccess
				if !hasAccess {
					log.Println("no access")
					rest.HttpErrorToJson(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
					return
				} else {
					next.ServeHTTP(w, r)
				}
			}				
		}
	}
}

