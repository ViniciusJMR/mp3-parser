# MP3 Decoder
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org/)

A simple MP3 frame parse library so you can have a structured MP3 frames. 

## Installation

```bash
go get github.com/ViniciusJMR/mp3-decoder
```

## Usage

```go
package main

import (
	"fmt"
	"github.com/ViniciusJMR/mp3-decoder"
)

func main() {
    f, err := os.Open("Luke-Bergs-Waesto-Follow-The-Sun(chosic.com).mp3")
	if err != nil {
		fmt.Printf("Error: %v", err)
	}

    frames, err := mp3.ParseFrames(f)

    for _, frame := range frames {
        processFrame(frame.ToBytes())
    }
}
```

## Roadmap
- [ ] Optimize memory usage on Frame Struct
- [ ] Add Mp3 Tag Support
- [ ] Add every MP3 Layer Support
