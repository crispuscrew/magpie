//go:build taskcheck

package notekeep

// No new API: db/export paths are already function parameters, so the ideal
// answer is (near) nothing. Gate = TestSelfcheck still passing; metrics tell the rest.
