package downloader

// SpeedTier represents a download speed tier configuration.
type SpeedTier string

const (
	// TierFast uses fewer concurrent files but more connections per file.
	TierFast SpeedTier = "fast"
	// TierStandard balances concurrent files and connections per file.
	TierStandard SpeedTier = "standard"
	// TierAuto automatically selects the best tier (defaults to standard).
	TierAuto SpeedTier = "auto"
)

// DownloadParams holds the tuned parameters for aria2c based on a speed tier.
type DownloadParams struct {
	ConcurrentFiles    int
	ConnectionsPerFile int
	PieceSize          string
	FileAllocation     string
}

// ParamsForTier returns the download parameters for the given speed tier string.
// Unrecognised tiers default to standard.
func ParamsForTier(tier string) DownloadParams {
	switch SpeedTier(tier) {
	case TierFast:
		return DownloadParams{
			ConcurrentFiles:    2,
			ConnectionsPerFile: 8,
			PieceSize:          "8M",
			FileAllocation:     "falloc",
		}
	default:
		return DownloadParams{
			ConcurrentFiles:    3,
			ConnectionsPerFile: 4,
			PieceSize:          "4M",
			FileAllocation:     "falloc",
		}
	}
}
