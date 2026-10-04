package parse

import "testing"

func TestNumberFromName(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		// baseline — must always work
		{"SSIS-001.mp4", "SSIS-001"},
		{"ABP-030-C", "ABP-030"},
		{"[Attackers] SHKD-474-C.mp4", "SHKD-474"},
		{"MIDV-100.mp4", "MIDV-100"},
		{"FC2-PPV-123456", "FC2-123456"},
		{"FC2_PPV_123456", "FC2-123456"},
		{"vema-181-4k-C", "vema-181"},
		{"Pono-n8877-C", "Pono-n8877"},
		{"caribbeancom-123456-789", "123456-789"},
		{"1Pondo-051920_123", "051920-123"},
		{"10musume-051920_123", "051920-123"},

		// user-reported failures (SDK gets these wrong)
		{"STARS-160_Uncensored_Leaked", "STARS-160"},
		{"MIMK-082_1080p_MR", "MIMK-082"},
		{"hdd600.com@MMGH-251_UNCENSORED_LEAKED_NOWATERMARK", "MMGH-251"},
		{"SNIS-326-uncensored-nyap2p.com", "SNIS-326"},
		{"SNIS-326-Uncen", "SNIS-326"},
		{"SNIS-326-Unc", "SNIS-326"},
		{"SNIS-326-U", "SNIS-326"},
		{"SNIS-326_UC", "SNIS-326"},
		{"SNIS-326-uc", "SNIS-326"},
		{"SNIS-326_uncensored", "SNIS-326"},

		// quality / source markers (separate from studio prefix)
		{"SNIS-326-FHD", "SNIS-326"},
		{"SNIS-326-RARBG", "SNIS-326"},
		{"SNIS-326-WEB-DL", "SNIS-326"},
		{"SNIS-326-REPACK", "SNIS-326"},

		// torrent site markers
		{"abc.nyap2p.com", ""},
		{"SNIS-326-nyaa", "SNIS-326"},
		{"SNIS-326-RARBG", "SNIS-326"},

		// user-reported v2 failures: SDK partial-matches Uncen / LEAKED /
		// NOWATERMARK, leaving stray letters in the number portion
		{"hdd600.com@MMGH-251_UNCENSORED_LEAKED_NOWATERMARK", "MMGH-251"},
		{"ipx-564-uncensored", "ipx-564"},
		{"IPX-506_Uncen", "IPX-506"},
		{"ABW-023_Uncen", "ABW-023"},
		{"hdd600.com@SDDE-618_UNCENSORED_LEAKED_NOWATERMARK", "SDDE-618"},

		// no-separator forms (metatube-server handles these too)
		{"SNIS326.mp4", "SNIS326"},
		{"SNIS326", "SNIS326"},
		{"SNIS326-Uncen", "SNIS326"},
		{"ABP030.mp4", "ABP030"},

		// garbage must yield ""
		{"random-junk.txt", ""},
		{"random-junk.mp4", ""},
		{"[some release group] totally not a number.mp4", ""},
	}
	for _, tc := range cases {
		got := NumberFromName(tc.in)
		if got != tc.want {
			t.Errorf("NumberFromName(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
