package auth

// tokenParams holds the resolved values from TokenOption functions
type tokenParams struct {
	subject  string
	audience []string
	admin    bool
	readOnly bool
}

// TokenOption is a functional option for configuring token generation
type TokenOption func(*tokenParams)

// WithSubject sets the subject claim on the token
func WithSubject(subject string) TokenOption {
	return func(p *tokenParams) {
		p.subject = subject
	}
}

// WithAudience sets the audience claim on the token
func WithAudience(audience ...string) TokenOption {
	return func(p *tokenParams) {
		p.audience = audience
	}
}

// WithAdmin sets the admin claim on the token
func WithAdmin() TokenOption {
	return func(p *tokenParams) {
		p.admin = true
	}
}

// WithReadOnly sets the read-only claim on the token
func WithReadOnly() TokenOption {
	return func(p *tokenParams) {
		p.readOnly = true
	}
}

func applyTokenOptions(opts []TokenOption) tokenParams {
	var p tokenParams
	for _, opt := range opts {
		opt(&p)
	}
	return p
}
