package auth

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWTAuth_Authenticate(t *testing.T) {
	b, _ := os.ReadFile("testdata/jwtRS256.key.pub")
	publicKey, _ := jwt.ParseRSAPublicKeyFromPEM(b)
	tests := []struct {
		name        string
		token       string
		key         interface{}
		options     []jwt.ParserOption
		want        bool
		wantSubject string
		wantAdmin   bool
	}{
		{
			name:        "valid RSA256 token",
			token:       "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0LXVzZXIiLCJhZG1pbiI6dHJ1ZX0.W4GpV3q33W7f2rUWolDSCC2y97UFweVCwAXoxVflOF4nKnXCMVkUIYYmZNs4_eVTTctta8soS1NcHsb4rN4ZvO5CSdrr5pPRV3ewoNXb0WmlNQz-9WGl5SgQddXw485SUQgQJr3J8lS5O9aFLUi0GEu9j3b85wrajVX_rFdF2JXCL7486uf8BDXXziJk_FwExLuK7S0iLPZPqLhcoeoPSZqf_2pZuKU_KpqLh7CM7yw_8gBL1YDRVHXJrortB34ip3-QZF8TuuTmMYPJWLpgxa2uIJF3XE9r207jxH-nSVqmbbIMeZBRKyN1CmNaAYQTZpOnQwaa86qM_ysrngslQtAbrfLswbFCIf3AiUe5GJgQRlZZlk_bvmJYc19KKn2ypLfNdZpqmlpnYH0oWKLIiaPWmrta4433BDeZK5SKrnqWLLivqkswEuPBO_xqhmdMdvhmETGlx2O8uubcUWFxz35h9T8ikzHCIP6Lxj_lGjLmTe02aOvMf9cii4atO9gEg6siu4N6Xf1Que6WGgdeR73UetbZ0rkPLmYTAfW72vpwM7_TlcMYZRLMz3rezXfVHkBCw8k0dri39hRg-XblS8S6Ij16UaRVG7TOp0pGO7aXKT5XjqVGX8L0dtiT_VRdEqXAIXekERGFAAjqq88Lu1l7IMMnUFYqvm-9ZHcfd40",
			key:         publicKey,
			want:        true,
			wantSubject: "test-user",
			wantAdmin:   true,
		},
		{
			name:  "invalid RSA256 token",
			token: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwiYWNjb3VudElEIjoidGVzdC1hY2NvdW50In0.GNbL5L2-UH37TeoyCe8SA-PPMLqlIYAmZwoJcEP7uyYmXulmDh_fOF2rHEN5ZP3dNw1PrJKEkOfQ2ebz820AOMjqvjI3zcLtuEA3srJcKcY_fv3EIXiLkrNDBt8tEiW-K2B8DKqe5xVI_I9uYkJCF9OfMNgVbqZ4lfq1KLSlBP_A9W2Y88syfwkEZo6lNl_WeDv9N2KkpMYG2pvfEYy6P238sTOld-BhGWq8ZO6bMsMtrxuahzIW-Zia5377HknbjdBkbmwdrikkN-ejD1VWA07Qj7w4TXkLkAP2xKP37gfKJYOU5HF_tSTwQ7bqJnvn3Ndg_avE658AoFTK_gbu0w",
			key:   publicKey,
			want:  false,
		},
		{
			name:  "valid RSA256 token with validation options",
			token: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.yRKA89MlgDVAViujkELt6uIgRHJPHpN_mpVPZhNiYe2L89977CUimQ3RzPnKqWr44d1mc0YeJf_suHWwwInJ7Fsn5INu5Q-Lv-L4BCg0Y8_OnsXNJPRYPPBN09zGf_88GoK0up6Y6QslkWBlv2YjRIVL5z9_OkX-arWco4tr2rQVcaP_7kUP89A25vjSTOP9lU8O5GQzLVg4LPpiE7XPSmadMHRQcNiQ-AfPpdj3h_uoTCcP8KwZtGa0gLCi-jWQKHeVW7tmXvplgqfNIMJnGNE3qW_ay_KyfLB5ooD0faKbOSEy-wegZk97DyZ80J9Ds-7c6smDsrBRDuhuLNjyXJdjmFWR7d_3-TkEZYVidsS_jcHnn3eL1OPmZE_bZMAaZ9XhWTkPrlktjd1uemcTM_OonV9-fXHKhpWMypzcKYTkRI_3oMT6sZKVQQziGesx_KAmPO7Fj9DxK74u7XMzstViUfmkmn7Yid-XPG6F1S-Pa-jfotmB3DJMgNlB6nMsTxdR9NcWibI4vvXc1Y_fBIV5era9vwzFkSy6NmO8Th8D-DLLXYJQT_5vH_56FLteVC-jPiwkXMUC6KVlgBqZ2V4I3uGiqGSWUVc9Hg1CDvQvSqJSdUybbv1iz079ZdMazZ3Mh3ePZfHj3vrYTbej-leJWRBA10cPpGeiPZ8YLGQ",
			options: []jwt.ParserOption{
				jwt.WithSubject("test-subject"),
			},
			key:  publicKey,
			want: false,
		},
		{
			name:        "valid HS256 token",
			token:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.drt_po6bHhDOF_FJEHTrK-KD8OGjseJZpHwHIgsnoTM",
			key:         []byte("mysecret"),
			wantSubject: "1234567890",
			want:        true,
		},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			a := &JWTAuth{AuthKey: tt.key, Options: tt.options}
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.Header.Set("Authorization", tt.token)
			ctx, auth := a.Authenticate(r)
			if auth != tt.want {
				t.Errorf("JWTAuth.Authenticate() auth = %v, want %v", auth, tt.want)
			}

			authorization, _ := ctx.Value(AuthorizationContextKey).(string)
			if auth != tt.want {
				t.Errorf("JWTAuth.Authenticate() ctx.AuthorizationContextKey = %v, want %v", authorization, tt.token)
			}

			subject, _ := ctx.Value(SubjectContextKey).(string)
			if subject != tt.wantSubject {
				t.Errorf("JWTAuth.Authenticate() ctx.SubjectContextKey = %v, want %v", subject, tt.wantSubject)
			}

			admin, _ := ctx.Value(AdminContextKey).(bool)
			if admin != tt.wantAdmin {
				t.Errorf("JWTAuth.Authenticate() ctx.AdminContextKey = %v, want %v", admin, tt.wantAdmin)
			}
		})
	}
}
