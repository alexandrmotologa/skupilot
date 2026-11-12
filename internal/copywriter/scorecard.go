package copywriter

import (
	"strings"

	"github.com/alexandrmotologa/skupilot/internal/domain"
)

// CalculateQualityScorecard computes the overall 0-100 listing quality metric.
func CalculateQualityScorecard(p *domain.Product) domain.QualityScorecard {
	card := domain.QualityScorecard{
		Recommendations: []string{},
	}

	// 1. Visual Compliance (30 Points Max)
	if p.ImageWhiteBGScore >= 0.985 {
		card.VisualScore += 15
	} else if p.ImageWhiteBGScore >= 0.900 {
		card.VisualScore += 8
		card.Recommendations = append(card.Recommendations, "Increase background pure whiteness to at least 98.5% (RGB 255)")
	} else {
		card.Recommendations = append(card.Recommendations, "Main image fails pure white background requirement (RGB 255)")
	}

	if p.ImageZoomReady {
		card.VisualScore += 10
	} else {
		card.Recommendations = append(card.Recommendations, "Upload image with minimum 1000x1000 pixels for marketplace zoom")
	}

	if !p.ImageWatermarkDetected {
		card.VisualScore += 5
	} else {
		card.Recommendations = append(card.Recommendations, "Remove watermark or promotional overlay from image")
	}

	// 2. Attribute Richness (30 Points Max)
	if p.ExtractedAttributes.Material != "" {
		card.AttributeScore += 8
	} else {
		card.Recommendations = append(card.Recommendations, "Specify exact fabric or material composition")
	}

	if p.ExtractedAttributes.Fit != "" {
		card.AttributeScore += 6
	} else {
		card.Recommendations = append(card.Recommendations, "Specify garment fit and silhouette")
	}

	if p.ExtractedAttributes.Collar != "" || p.ExtractedAttributes.Neckline != "" {
		card.AttributeScore += 6
	}

	if p.ExtractedAttributes.ColorFamily != "" || p.ExtractedAttributes.ColorHex != "" {
		card.AttributeScore += 5
	}

	if p.ExtractedAttributes.Closure != "" || p.ExtractedAttributes.Pattern != "" {
		card.AttributeScore += 5
	}

	// 3. Title Quality (20 Points Max)
	titleLen := len(p.OptimizedTitle)
	if titleLen >= 140 && titleLen <= 200 {
		card.TitleScore += 10
	} else if titleLen >= 80 {
		card.TitleScore += 6
		card.Recommendations = append(card.Recommendations, "Expand title toward 150-180 characters for optimal Amazon search indexing")
	} else {
		card.Recommendations = append(card.Recommendations, "Title is too short for marketplace indexing")
	}

	if p.Brand != "" && strings.Contains(strings.ToLower(p.OptimizedTitle), strings.ToLower(p.Brand)) {
		card.TitleScore += 5
	}

	hasBanned := false
	lowerTitle := strings.ToLower(p.OptimizedTitle)
	for _, b := range bannedPromotionalPhrases {
		if strings.Contains(lowerTitle, b) {
			hasBanned = true
			break
		}
	}
	if !hasBanned {
		card.TitleScore += 5
	} else {
		card.Recommendations = append(card.Recommendations, "Remove promotional words (e.g. sale, best seller) from title")
	}

	// 4. Bullet Points Quality (20 Points Max)
	if len(p.BulletPoints) == 5 {
		card.BulletScore += 10
	} else {
		card.Recommendations = append(card.Recommendations, "Listing requires exactly 5 benefit-driven feature bullets")
	}

	allBulletsWellFormed := true
	for _, b := range p.BulletPoints {
		if len(b) < 60 || !strings.Contains(b, ":") {
			allBulletsWellFormed = false
			break
		}
	}
	if allBulletsWellFormed && len(p.BulletPoints) > 0 {
		card.BulletScore += 10
	} else {
		card.BulletScore += 5
	}

	card.TotalScore = card.VisualScore + card.AttributeScore + card.TitleScore + card.BulletScore
	if card.TotalScore > 100 {
		card.TotalScore = 100
	}

	return card
}
