// Package parse extracts an AV番号 from a filename or folder name.
//
// Strategy: rather than rely on the SDK's lossy regex (which leaves
// stray letters like "STARS-160_Uncensored_Leaked" → "STARS-160e"),
// we:
//
//  1. Strip extension and SDK-style decoration.
//  2. Split into tokens by separator chars (- _ . space @).
//  3. Find the first letter-only short token (likely the prefix) and
//     the first digit-only short token (likely the number), then
//     glue them as PREFIX-NUMBER.
//
// We deliberately do NOT use the SDK's number.Trim output as the
// final answer because its regex chain is too greedy — it partially
// matches release-tag fragments and leaves stray characters. Instead
// we use the SDK for two things only:
//   - bracket/domain stripping in the initial pass
//   - FC2 PPV normalization (FC2_PPV_123456 → FC2-123456)
//
// Recognized番号 shapes:
//   - PREFIX-NUMBER     e.g. SNIS-326, STARS-160, vema-181
//   - NUMBER-NUMBER     e.g. 1234-56, 051920-123
//   - FC2-NUMBER        e.g. FC2-1234567
//   - PURE-DIGIT (>=5)  e.g. 051920_123 — accepted as last fallback
package parse

import (
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/metatube-community/metatube-sdk-go/common/number"
)

// shortLetters is 1-8 ASCII letters, used to identify a token that
// might be a番号 prefix (e.g. SNIS, STARS, MIDV, vema, n8877).
var shortLetters = regexp.MustCompile(`^[A-Za-z]{1,8}$`)

// shortDigits is 1-6 ASCII digits, used to identify a token that might
// be a番号 number portion.
var shortDigits = regexp.MustCompile(`^\d{1,6}$`)

// prefixNumberRe is the canonical "PREFIX-NUMBER" shape, used for
// whole-string validation AFTER the SDK has cleaned up most junk.
// PREFIX is uppercase only (most AV番号 use uppercase; lowercase
// prefixes like "vema" only appear in the per-token path).
var prefixNumberRe = regexp.MustCompile(`^[A-Z]{1,8}-\d{2,6}$`)

// prefixNumberMixedRe accepts mixed-case prefix tokens ("n8877",
// "vema181" without a separator). Used by strategies 1b and 4.
var prefixNumberMixedRe = regexp.MustCompile(`^[A-Za-z]{1,8}\d{1,6}$`)

// NumberFromName returns the canonical番号 extracted from name, or ""
// if no plausible番号 can be found. Folder names work the same way —
// we strip the path and trim the extension (which folders won't have,
// but the trim function tolerates that).
func NumberFromName(name string) string {
	base := filepath.Base(name)
	// IMPORTANT: use path.Ext (not filepath.Ext). path.Ext only treats
	// the trailing segment as an extension if it's short (≤7 chars),
	// so ".com@MMGH-251_test" is NOT mistaken for an extension. This
	// matters for filenames like "hdd600.com@MMGH-251..." where
	// filepath.Ext would greedily strip the whole rest of the string.
	if ext := path.Ext(base); len(ext) > 0 && len(ext) <= 7 {
		base = strings.TrimSuffix(base, ext)
	}
	if base == "" {
		return ""
	}

	// Fast path: if the SDK's whole-string Trim happens to produce a
	// clean番号 (the easy cases), use it directly. This covers
	// "SSIS-001.mp4", "ABP-030-C", "vema-181-4k-C" etc.
	if trimmed := number.Trim(base); trimmed != "" {
		if prefixNumberRe.MatchString(trimmed) ||
			strings.HasPrefix(trimmed, "FC2-") ||
			digitPairRe.MatchString(trimmed) {
			return strings.ReplaceAll(trimmed, "_", "-")
		}
	}

	// Fall back to per-token scanning. The SDK's regex chain is too
	// greedy to trust for messy filenames (e.g. "hdd600.com@..." gets
	// truncated to "hdd600" because the regex thinks "hd" is an "HD"
	// release prefix), so we only use it for the fast path above and
	// rely on token matching for everything else.
	tokens := filter(splitTokens(base))

	// Strategy 1: first letter-only token followed by first
	// digit-or-digit-mixed token. Handles "STARS-160_Uncensored_Leaked"
	// → split → ["STARS", "160", "Uncensored", "Leaked"] → STARS-160.
	// Also handles "Pono-n8877-C" → ["Pono", "n8877", "C"] → fails on
	// strategy 1 because "n8877" is mixed, then strategy 1b below
	// finds "n8877" as a prefix-mixed token.
	for _, tok := range tokens {
		if !shortLetters.MatchString(tok) {
			continue
		}
		idx := indexOf(tokens, tok)
		if idx < 0 || idx == len(tokens)-1 {
			break
		}
		for _, ntok := range tokens[idx+1:] {
			if shortDigits.MatchString(ntok) {
				return tok + "-" + ntok
			}
			if isJunkToken(ntok) {
				continue
			}
			if len(ntok) > 8 {
				continue
			}
		}
		break
	}
	// letter+digit mixed token (handles n8877, vema-style番号 where
	// the number portion has a leading letter).
	for _, tok := range tokens {
		if !shortLetters.MatchString(tok) {
			continue
		}
		idx := indexOf(tokens, tok)
		if idx < 0 || idx == len(tokens)-1 {
			break
		}
		for _, ntok := range tokens[idx+1:] {
			if prefixNumberMixedRe.MatchString(ntok) {
				return tok + "-" + ntok
			}
			if isJunkToken(ntok) {
				continue
			}
			if len(ntok) > 10 {
				continue
			}
		}
		break
	}

	// Strategy 2: FC2. The SDK normalizes "FC2_PPV_1234567" to
	// "FC2-1234567"; we just check the IsFC2 gate on any token.
	for _, tok := range tokens {
		if number.IsFC2(number.Trim(tok)) {
			return number.Trim(tok)
		}
	}

	// Strategy 3: pure-digit pair (Caribbeancom 1234-56,
	// 1Pondo 051920_123).
	for i := 0; i < len(tokens)-1; i++ {
		if shortDigits.MatchString(tokens[i]) &&
			shortDigits.MatchString(tokens[i+1]) {
			return tokens[i] + "-" + tokens[i+1]
		}
	}

	// Strategy 4: any single token that already looks like a clean
	// mixed-case番号 ("n8877", "vema181" if no separator).
	for _, tok := range tokens {
		if prefixNumberMixedRe.MatchString(tok) {
			return tok
		}
	}

	// Strategy 5: scan the whole base for any "LETTERSDIGITS" pattern.
	// Handles filenames like "SNIS326.mp4" or "SNIS326-Uncen" where
	// the prefix and number are concatenated without a separator. The
	// SDK's number.Trim doesn't catch this case; we look for any
	// 2-8 ASCII letters immediately followed by 2-6 ASCII digits
	// anywhere in the base.
	for _, m := range noSepNumberRe.FindAllString(base, -1) {
		return strings.ToUpper(m)
	}

	return ""
}

// noSepNumberRe matches a prefix + digits with no separator between
// them. Used by strategy 5 to recover "SNIS326" from "SNIS326.mp4"
// or "SNIS326-Uncen.mp4" when token splitting missed it.
var noSepNumberRe = regexp.MustCompile(`[A-Za-z]{2,8}\d{2,6}`)

// digitPairRe matches "NUMBER-NUMBER" / "NUMBER_NUMBER" — used by the
// whole-string SDK-trim path above.
var digitPairRe = regexp.MustCompile(`^\d{3,6}[-_]\d{1,6}$`)

// splitTokens splits on separator chars and "@" (with the right-hand
// side preferred). See NumberFromName comment for the @ rationale.
func splitTokens(s string) []string {
	tokens := splitOnSeparators(s, "-_. ")
	out := make([]string, 0, len(tokens))
	for _, tok := range tokens {
		if i := strings.Index(tok, "@"); i >= 0 {
			right := tok[i+1:]
			left := tok[:i]
			if right != "" {
				out = append(out, splitOnSeparators(right, "-_. ")...)
			}
			if left != "" {
				out = append(out, splitOnSeparators(left, "-_. ")...)
			}
		} else {
			out = append(out, tok)
		}
	}
	return out
}

func splitOnSeparators(s, seps string) []string {
	return strings.FieldsFunc(s, func(r rune) bool {
		return strings.ContainsRune(seps, r)
	})
}

// filter removes empty strings, trivially-short tokens, and known
// release-group / site names that aren't part of a番号. We do NOT
// strip studio names like "1Pondo" / "caribbeancom" / "10musume"
// here — those are handled by the SDK's number.Trim regex on the
// fast-path branch above.
func filter(tokens []string) []string {
	out := make([]string, 0, len(tokens))
	for _, tok := range tokens {
		tok = strings.TrimSpace(tok)
		if tok == "" || len(tok) < 2 {
			continue
		}
		if isReleaseGroupToken(tok) {
			continue
		}
		out = append(out, tok)
	}
	return out
}

// isReleaseGroupToken returns true for tokens that are known
// torrent-site / release-group names. These are NEVER part of a valid
//番号 — they're always junk that appeared in the filename from the
// torrent site where the user downloaded it.
func isReleaseGroupToken(s string) bool {
	switch strings.ToLower(s) {
	case "hdd600", "hdd800", "nyaa", "nya",
		"rarbg", "yify", "yts", "ith", "tlf", "thz",
		"xdsd", "xm", "xiu", "iqq",
		"sukebei", "tokyotosho", "anidex":
		return true
	}
	return false
}

// indexOf returns the index of s in tokens, or -1.
func indexOf(tokens []string, s string) int {
	for i, t := range tokens {
		if t == s {
			return i
		}
	}
	return -1
}

// isJunkToken returns true for tokens that are obviously release-group
// domains or TLDs, which can appear between the prefix and the number
// portion of a番号 (e.g. "hdd600.com@MMGH-251" splits to
// ["hdd600", "MMGH", "com", "251"]; "com" is junk between MMGH and 251).
func isJunkToken(s string) bool {
	switch strings.ToLower(s) {
	case "com", "net", "org", "io", "tv", "to", "co", "jp", "cn":
		return true
	}
	return false
}
