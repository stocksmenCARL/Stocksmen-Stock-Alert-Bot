/*
 * Copyright 2022 M1K
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package screenshot

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/google/uuid"
)

func TakeSS(ticker string) string {

	// Start Chrome
	// Remove the 2nd param if you don't need debug information logged
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	img, _ := filepath.Abs("./" + ticker + ".html")
	defer func() {
		os.Remove(img)
	}()

	url := "file:///" + img
	filename := fmt.Sprintf("%v.png", uuid.New().String())

	// Run Tasks
	// List of actions to run in sequence (which also fills our image buffer)
	var imageBuf []byte
	if err := chromedp.Run(ctx, ScreenshotTasks(url, &imageBuf)); err != nil {
		log.Fatal(err)
	}

	// Write our image to file
	if err := ioutil.WriteFile(filename, imageBuf, 0644); err != nil {
		log.Fatal(err)
	}

	return filename
}

func ScreenshotTasks(url string, imageBuf *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.ActionFunc(func(ctx context.Context) (err error) {
			*imageBuf, err = page.CaptureScreenshot().WithQuality(100).WithClip(&page.Viewport{
				X:      5,
				Y:      5,
				Width:  750,
				Height: 465,
				Scale:  1.0,
			}).Do(ctx)
			return err
		}),
	}
}
