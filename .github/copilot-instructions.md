# Coding Guidelines

## Code Comment Guidelines

We prefer concise, meaningful comments that explain "why" rather than "what". Follow these guidelines:

### Good Comments

- Comment on complex business logic explaining the "why" behind it
- Document non-obvious edge cases or requirements
- Explain unusual approaches or performance optimizations
- Document interfaces with brief method descriptions
- Use helpful godoc comments for exported functions/types

### Avoid

- Comments that simply repeat what the code does
- Unnecessary comments on trivial code
- Commented-out code (use version control instead)
- Excessive inline comments for basic operations
- Auto-generated comment blocks that don't add value

## GO Code Style Guidelines

### Error Handling

```go
// Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to create client: %w", err)
}
```

### Logging

Use structured logging with `slog` package and always use the "err" attribute for errors:

```go
slog.InfoContext(ctx, "processing request", "user_id", userID)
slog.ErrorContext(ctx, "failed to create client", "err", err)
```

### Testing

Use table-driven tests with testify/require for assertions,
Use `t.Context()` for passing context in tests,
Use `gomock` with a mockSetup test case function for mocking dependencies in tests:
For hardcoded identifiers use the format `test-{name type}` ex: `test-user`, `test-error`
  When there are multiple identifiers use the format `test-{name type}-{index}` ex: `test-user-1`, `test-user-2`

```go
func TestService_DoSomething(t *testing.T) {
    tests := []struct {
        name      string
        input     string
        mockSetup func(mockClient *mock.MockClient)
        expected  string
        wantError bool
    }{
        // Test cases
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctrl := gomock.NewController(t)
            defer ctrl.Finish() 
            mockClient := mock.NewMockClient(ctrl)
            t.mockSetup(mockClient)

            s := TestService{
                client: mockClient,
            }

            result, err := s.DoSomething(t.Context(), tt.input)
            
            require.Equal(t, tt.expected, result)
            require.Equal(t, tt.wantError, err != nil)
        })
    }
}
```


### Service Implementation

Services follow dependency injection pattern with interfaces for testability:

```go
type Service interface {
    DoSomething(ctx context.Context, input string) (string, error)
}

type service struct {
    client Client
    db     Database
}

func NewService(client Client, db Database) Service {
    return &service{client: client, db: db}
}
```

