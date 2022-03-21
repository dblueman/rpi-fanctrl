package main

import (
   "fmt"
   "os"
   "strconv"
   "strings"
   "time"

   "github.com/felixge/pidctrl"
)

var (
   inPath  string = "/sys/devices/virtual/thermal/thermal_zone0/hwmon0/temp1_input"
   outPath string = "/sys/devices/platform/rpi-poe-fan@0/hwmon/hwmon1/pwm1"
   control string = "/sys/class/thermal/thermal_zone0/mode"
   quit    bool
)

const (
   min      float64 = 0.
   max      float64 = 255.
   setpoint float64 = 55.
   p        float64 = -8.
   i        float64 = -0.1
   d        float64 = 0.
   window   int     = 5
)

func override(mode string) {
   err := os.WriteFile(control, []byte(mode), 0o666)
   if err != nil {
      panic(err)
   }
}

func main() {
   pid := pidctrl.NewPIDController(p, i, d)
   pid.SetOutputLimits(min, max)
   pid.Set(setpoint)

   override("disabled")
   defer override("enabled")

   outStr, err := os.ReadFile(outPath)
   if err != nil {
      fmt.Println(err)
      return
   }

   outLast, err := strconv.Atoi(strings.TrimSpace(string(outStr)))
   if err != nil {
      fmt.Println(err)
      return
   }

   for !quit {
      var inTotal float64

      for i := 0; i < window; i++ {
         time.Sleep(100 * time.Millisecond)

         var inStr []byte
         inStr, err = os.ReadFile(inPath)
         if err != nil {
            fmt.Println(err)
            return
         }

         inStr2 := strings.TrimSpace(string(inStr))

         var in float64
         in, err = strconv.ParseFloat(inStr2, 64)
         if err != nil {
            fmt.Println(err)
            return
         }

         inTotal += in
      }

      inTotal /= float64(window) * 1000.
      out := int(pid.Update(inTotal))

      if out == outLast {
         continue
      }

      outStr := strconv.Itoa(out)

      err = os.WriteFile(outPath, []byte(outStr), 0o666)
      if err != nil {
         fmt.Println(err)
         return
      }

      outLast = out
   }
}

