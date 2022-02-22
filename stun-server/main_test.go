func TestValidateRequest(t *testing.T) {
    type test struct {
        input []byte
        sep   string
        want  []string
    }

    tests := []test{
        {input: make([]byte, 128), want: fmt.Errorf("Request is type %x, should be Binding Request (0x0001)", messageType),
        {input: make([]byte, 128), want: fmt.Errorf("Request is %d bytes, should be 20", 1),
    }

    for _, tc := range tests {
        got := validateRequest(tc.input, make([]byte, 128))
        if !reflect.DeepEqual(tc.want, got) {
            t.Fatalf("expected: %v, got: %v", tc.want, got)
        }
    }
}