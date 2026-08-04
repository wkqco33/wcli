package rich

// DisplayWidth는 문자열이 터미널에서 차지하는 칸 수를 반환합니다.
// 동아시아 전각 문자(한글/한자/가나/전각기호)와 주요 이모지는 2칸,
// 그 외는 1칸으로 계산합니다.
//
// ANSI 코드나 마크업 태그는 고려하지 않으므로, 폭을 정확히 재려면
// stripMarkup 등으로 태그를 먼저 제거한 뒤 호출해야 합니다.
func DisplayWidth(s string) int {
	w := 0
	for _, r := range s {
		w += runeWidth(r)
	}
	return w
}

// runeWidth는 룬 하나의 표시 폭(1 또는 2)을 반환합니다.
func runeWidth(r rune) int {
	if isWide(r) {
		return 2
	}
	return 1
}

// isWide는 전각(폭 2)으로 표시되는 문자인지 판별합니다.
// 한글/한자/가나/전각기호 및 주요 이모지 범위를 포함합니다.
func isWide(r rune) bool {
	switch {
	case r >= 0x1100 && r <= 0x115F, // 한글 자모
		r >= 0x2E80 && r <= 0x303E,   // CJK 부수/기호
		r >= 0x3041 && r <= 0x33FF,   // 히라가나/가타카나/한중일 기호
		r >= 0x3400 && r <= 0x4DBF,   // CJK 확장 A
		r >= 0x4E00 && r <= 0x9FFF,   // CJK 통합 한자
		r >= 0xA000 && r <= 0xA4CF,   // 이족 음절
		r >= 0xAC00 && r <= 0xD7A3,   // 한글 음절
		r >= 0xF900 && r <= 0xFAFF,   // CJK 호환 한자
		r >= 0xFE30 && r <= 0xFE4F,   // CJK 호환 형태
		r >= 0xFF00 && r <= 0xFF60,   // 전각 형태
		r >= 0xFFE0 && r <= 0xFFE6,   // 전각 기호
		r >= 0x1F000 && r <= 0x1FAFF, // 이모지
		r >= 0x2600 && r <= 0x27BF:   // 기타 기호/딩뱃
		return true
	}
	return false
}
