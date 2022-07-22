package sample

import (
	"math/rand"
	"time"

	"github.com/google/uuid"

	"laptop-app-using-grpc/pb/pb"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func randomKeyboardLayout() pb.Keyboard_Layout {
	switch rand.Intn(3) {
	case 1:
		return pb.Keyboard_QWERTY
	case 2:
		return pb.Keyboard_QWERTZ
	default:
		return pb.Keyboard_AZERTY
	}
}

func randomBool() bool {
	return rand.Intn(2) == 1
}

func randomStringFromSet(a ...string) string {
	n := len(a)
	if n == 0 {
		return ""
	}
	return a[rand.Intn(n)]
}

func randomCPUBrand() string {
	return randomStringFromSet("Intel", "AMD")
}

func randomCPUName(brand string) string {
	if brand == "Intel" {
		return randomStringFromSet(
			"Xeon E-2286M",
			"Core i9-9980HK",
			"Core i7-9750H",
			"Core i5-9400F",
			"Core i3-1005G1",
		)
	}

	return randomStringFromSet(
		"Ryzen 7 PRO 2700U",
		"Ryzen 5 PRO 3500U",
		"Ryzen 3 PRO 3200GE",
	)
}

func randomInt(min, max int) int {
	return min + rand.Int()%(max-min+1)
}

func randomFloat64(min float64, max float64) float64 {
	//Logic for generating random between min and max
	return min + rand.Float64()*(max-min)
}

func randomFloat32(min float32, max float32) float32 {
	//Logic for generating random between min and max
	return min + rand.Float32()*(max-min)
}

func randomGPUBrand() string {
	return randomStringFromSet("NVIDIA", "AMD")
}

func randomGPUName(brand string) string {
	if brand == "NVIDIA" {
		return randomStringFromSet(
			"RTX 2060",
			"RTX 2070",
			"RTX 2080-Ti",
			"GTX 1660-Ti",
		)
	}

	return randomStringFromSet(
		"RX 590",
		"RX 580",
		"RX 5700-XT",
		"RX Vega-56",
	)
}

func randomScreenPanel() pb.Screen_Panel {
	if rand.Intn(2) == 1 {
		return pb.Screen_IPS
	}

	return pb.Screen_OLED
}

func randID() string {
	return uuid.New().String()
}

func randomLaptopBrand() string {
	return randomStringFromSet("Apple", "Dell", "Lenovo", "Microsoft", "Acer", "HP")
}

func randomLaptopName(brand string) string {
	if brand == "Apple" {
		return randomStringFromSet(
			"MacBook Pro",
			"MacBook Air",
			"MacBook",
		)
	}

	if brand == "Dell" {
		return randomStringFromSet(
			"G15-5525",
			"Alienware x15",
			"Inspiron 3511",
			"Vostro 15 3568",
		)
	}

	if brand == "Lenovo" {
		return randomStringFromSet(
			"IdeaPad Slim 5",
			"Yoga 9i",
			"ThinkPad X13",
			"Legion 5 Pro",
		)
	}

	if brand == "Microsoft" {
		return randomStringFromSet(
			"Surface Pro 7",
			"Surface Book 3",
			"Surface Laptop 4",
			"Surface StudioBook",
		)
	}

	if brand == "HP" {
		return randomStringFromSet(
			"Envy 13",
			"Omen 16",
			"Spectre x360",
			"Pavilion 15s",
		)
	}

	return randomStringFromSet(
		"Aspire 7",
		"Swift 3",
		"Nitro 5",
		"Predator Helios",
	)
}
