package watchmen

func calculateCoverage(data []float64, compareValue float64) float64 {
	dataCount := len(data)
	var compareResult []float64
	for _, v := range data {
		if v > compareValue {
			compareResult = append(compareResult, v)
		}
	}
	compareResultCount := len(compareResult)
	return float64(compareResultCount) / float64(dataCount)

}

func linearRegression(x []float64, y []float64) (slope float64, intercept float64) {
	var (
		n            float64
		sumX, sumY   float64
		sumXY, sumX2 float64
	)

	n = float64(len(x))
	for i := 0; i < len(x); i++{
		sumX += x[i]
		sumX2 += x[i] * x[i]
		sumY += y[i]
		sumXY += x[i] * y[i]
	}
	covXY := sumXY - sumX*sumY/n
	varX := sumX2 - sumX*sumX/n

	slope = covXY / varX
	intercept = sumY/n - slope*sumX/n

	return
}

func driftXLinearRegression(x []float64, y []float64, offsetX float64) (slope float64, intercept float64) {
	var newX []float64
	for _, v := range x {
		newX = append(newX, v - offsetX)
	}
	return linearRegression(newX, y)
}

