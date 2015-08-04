// 用于生成Android studio的目录结构

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/template"
)

const (
	RAW_MANIFEST = `<?xml version="1.0" ?>
<manifest xmlns:android="http://schemas.android.com/apk/res/android"
          package="{{.Pkg}}">
    <application android:allowBackup="true"
        >

    </application>
</manifest>
`

	PLUGIN_LIBRARY     = "com.android.library"
	PLUGIN_APPLICATION = "com.android.application"

	RAW_GRADLE_BUILD = `
apply plugin: '{{.Plugin}}'

android{
	compileSdkVersion 22
	buildToolsVersion '22.0.1'

	defaultConfig{
		minSdkVersion 8
		targetSdkVersion 22
		versionCode 1
		versionName '1.0'
	}

	buildTypes{
		debug{
			minifyEnabled false
			proguardFiles getDefaultProguardFile('proguard-android.txt'), 'proguard-rules.pro'
		}
	}

}

dependencies{
	compile 'com.android.support:support-v4:22.2.0'
}

`

	RAW_PROGUARD = `# Add project specific ProGuard rules here.
# By default, the flags in this file are appended to flags specified
# in /home/zhou/opt/android-sdk/tools/proguard/proguard-android.txt
# You can edit the include path and order by changing the proguardFiles
# directive in build.gradle.
#
# For more details, see
#   http://developer.android.com/guide/developing/tools/proguard.html

# Add any project specific keep options here:

# If your project uses WebView with JS, uncomment the following
# and specify the fully qualified class name to the JavaScript interface
# class:
#-keepclassmembers class fqcn.of.javascript.interface.for.webview {
#   public *;
#}
`
)

// App
// ├── libs
// ├── src
// │   ├── androidTest
// │   │   └── java
// │   └── main
// │       ├── assets
// │       ├── java
// │       ├── jniLibs
// │       ├── res
// │       └── AndroidManifest.xml
// ├── build.gradle
// ├── proguard-rules.pro
// └── settings.gradle
// Studio的文件结构
type Studio struct {
	Libs          string
	Src           string
	Test          string
	Assets        string
	JniLibs       string
	Res           string
	Manifest      string
	GradleBuild   string
	GradleSetting string
	Proguard      string
	Pkg           string
	IsApp         bool
}

func NewStudio() *Studio {
	return &Studio{
		Libs:          "libs",
		Src:           "src/main/java",
		Test:          "src/androidTest/java",
		Assets:        "src/main/assets",
		JniLibs:       "src/main/jniLibs",
		Res:           "src/main/res",
		Manifest:      "src/main/AndroidManifest.xml",
		GradleBuild:   "build.gradle",
		GradleSetting: "settings.gradle",
		Proguard:      "proguard-rules.pro",
	}
}

func NewStudioWithBase(base string) *Studio {
	if base != "" {
		return &Studio{
			Libs:          base + "/libs",
			Src:           base + "/src/main/java",
			Test:          base + "/src/androidTest/java",
			Assets:        base + "/src/main/assets",
			JniLibs:       base + "/src/main/jniLibs",
			Res:           base + "/src/main/res",
			Manifest:      base + "/src/main/AndroidManifest.xml",
			GradleBuild:   base + "/build.gradle",
			GradleSetting: base + "/settings.gradle",
			Proguard:      base + "/proguard-rules.pro",
		}
	}
	return NewStudio()
}

// 生成目录结构
func (s *Studio) mkDirs() {
	// 生成文件夹目录
	mkDirs(s.Libs)
	mkDirs(s.Src)
	mkDirs(s.Test)
	mkDirs(s.Assets)
	mkDirs(s.JniLibs)
	mkDirs(s.Res)
	// 生成Manifest
	templ := template.New("manifest")
	templ.Parse(RAW_MANIFEST)
	manifest, err := os.Create(s.Manifest)
	if err != nil {
		fmt.Println("open manifest file error", err)
		return
	}
	defer manifest.Close()
	if s.Pkg != "" {
		fmt.Println("Package name is", s.Pkg)
	} else {
		fmt.Println("Warning! use a blank package name")
	}
	err = templ.Execute(manifest, map[string]string{
		"Pkg": s.Pkg,
	})
	if err != nil {
		fmt.Println("writer manifest file by template error", err)
		return
	}
	// 生成gradle
	templ = template.New("gradle")
	templ.Parse(RAW_GRADLE_BUILD)
	gradleBuild, err := os.Create(s.GradleBuild)
	if err != nil {
		fmt.Println("open build.gradle file error", err)
		return
	}
	defer gradleBuild.Close()
	var plugin string
	if s.IsApp {
		plugin = PLUGIN_APPLICATION
	} else {
		plugin = PLUGIN_LIBRARY
	}
	fmt.Println("use gradle plugin", plugin)
	err = templ.Execute(gradleBuild, map[string]string{
		"Plugin": plugin,
	})
	if err != nil {
		fmt.Println("writer build.gradle fail.", err)
		return
	}
	gradleSettings, err := os.Create(s.GradleSetting)
	if err != nil {
		fmt.Println("open settings.gradle fail")
		return
	}
	defer gradleSettings.Close()
	fmt.Fprint(gradleSettings, "\n")
	// 生成proguard
	proguard, err := os.Create(s.Proguard)
	if err != nil {
		fmt.Println("open proguard error.", err)
		return
	}
	defer proguard.Close()
	proguard.WriteString(RAW_PROGUARD)

	// copy java
	// copy assets
	// copy res
	// copy libs
}

// 只创建文件目录, 如果有扩展名,就只创建父目录
func mkDirs(path string) (err error) {
	dot := strings.LastIndex(path, ".")
	if dot > -1 {
		// 文件
		d := strings.LastIndex(path, "/")
		if d == -1 {
			return errors.New("no found folder in path, may only a file in root." + path)
		}
		parent := path[:d]
		return os.MkdirAll(parent, 0775)
	} else {
		return os.MkdirAll(path+"/", 0775)
	}
}

func main() {
	//mkDirs("path")
	var pkg string
	var isApp bool
	flag.StringVar(&pkg, "p", "", "-p {pkg}")
	flag.BoolVar(&isApp, "app", false, "[-app] to mark the project as Application. default is Library")
	flag.Parse()
	baseDir := flag.Arg(0)
	studio := NewStudioWithBase(baseDir)
	studio.Pkg = pkg
	studio.IsApp = isApp
	studio.mkDirs()
}
