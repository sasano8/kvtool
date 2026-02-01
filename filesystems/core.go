package filesystems

type MountConfig struct {
	Dir  *string `json:"dir"`
	File *string `json:"file"`
}

type TransformConfig struct {
	Read  string `json:"read"`
	Write string `json:"write"`
}

type StoreConfig struct {
	Driver string         `json:"version"`
	Args   map[string]any `json:"args"`
	Mount  *MountConfig   `json:"mount"`
}

// ".env":
//       driver: "local"
//       args:
//         ext: ".env"
//         transform:
//           read: dotenv
//           write: dotenv
//         root: "."
//       mount:
//         # dir: ""
//         file: ""
