package charter

import (
  "fmt"
  "os"
  "github.com/go-echarts/go-echarts/v2/charts"
  "github.com/go-echarts/go-echarts/v2/opts"
  "github.com/go-echarts/go-echarts/v2/types"
)

func MakeChart(seriesName string, seriesIndex int, data []float64) {
  items := make([]opts.LineData, len(data), len(data))

  xLabels := make([]string, len(data), len(data))

  for i := 0; i < len(data); i++ {
    items[i] = opts.LineData{
      Value: data[i],
    }
    xLabels[i] = fmt.Sprint(i)
  }

  line := charts.NewLine()
  line.SetGlobalOptions(
    charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeWesteros}),
    charts.WithTitleOpts(opts.Title{
      Title:    seriesName,
      Subtitle: string(seriesIndex),
    }),
  )

  line.SetXAxis(xLabels).
    AddSeries("Data", items).
    SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: false}))

  f, err := os.Create(fmt.Sprintf("/tmp/charter/" + seriesName + "_%d.html", seriesIndex))

  if err != nil {
    panic(err)
  }
  defer f.Close()

  line.Render(f)
}
