package downloader

import "testing"

func TestSpeedTierParams(t *testing.T) {
	tests := []struct {
		tier   string
		expect DownloadParams
	}{
		{
			tier: "fast",
			expect: DownloadParams{
				ConcurrentFiles:    2,
				ConnectionsPerFile: 8,
				PieceSize:          "8M",
				FileAllocation:     "falloc",
			},
		},
		{
			tier: "standard",
			expect: DownloadParams{
				ConcurrentFiles:    3,
				ConnectionsPerFile: 4,
				PieceSize:          "4M",
				FileAllocation:     "falloc",
			},
		},
		{
			tier: "auto",
			expect: DownloadParams{
				ConcurrentFiles:    3,
				ConnectionsPerFile: 4,
				PieceSize:          "4M",
				FileAllocation:     "falloc",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.tier, func(t *testing.T) {
			got := ParamsForTier(tt.tier)
			if got != tt.expect {
				t.Errorf("ParamsForTier(%q) = %+v, want %+v", tt.tier, got, tt.expect)
			}
		})
	}
}
