// Prevents additional console window on Windows in release, DO NOT REMOVE!!
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

use tauri::{
    menu::{Menu, MenuItem},
    tray::{MouseButton, MouseButtonState, TrayIconBuilder, TrayIconEvent},
    Manager,
};

#[cfg(target_os = "windows")]
use std::os::windows::process::CommandExt;

#[cfg(target_os = "windows")]
fn get_windows_user_shell_dirs(app: &tauri::AppHandle) -> Vec<std::path::PathBuf> {
    let mut dirs: Vec<std::path::PathBuf> = Vec::new();

    // 1. Query registry for true redirected folders (e.g. D:\Users\Lenovo\Downloads, D:\Users\Lenovo\Desktop)
    if let Ok(output) = std::process::Command::new("reg")
        .args(&["query", "HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Explorer\\User Shell Folders"])
        .output()
    {
        let stdout = String::from_utf8_lossy(&output.stdout);
        for line in stdout.lines() {
            if let Some(idx) = line.find("REG_EXPAND_SZ").or_else(|| line.find("REG_SZ")) {
                let val = line[idx + 13..].trim();
                let expanded = if val.contains("%USERPROFILE%") {
                    if let Ok(up) = std::env::var("USERPROFILE") {
                        val.replace("%USERPROFILE%", &up)
                    } else {
                        val.to_string()
                    }
                } else {
                    val.to_string()
                };
                let p = std::path::PathBuf::from(expanded);
                if p.exists() && !dirs.contains(&p) {
                    dirs.push(p);
                }
            }
        }
    }

    // 2. Tauri path API complements
    if let Ok(p) = app.path().download_dir() { if p.exists() && !dirs.contains(&p) { dirs.push(p); } }
    if let Ok(p) = app.path().desktop_dir() { if p.exists() && !dirs.contains(&p) { dirs.push(p); } }
    if let Ok(p) = app.path().document_dir() { if p.exists() && !dirs.contains(&p) { dirs.push(p); } }

    // 3. Fallback standard USERPROFILE paths
    if let Ok(userprofile) = std::env::var("USERPROFILE") {
        for sub in &["Downloads", "Desktop", "Documents"] {
            let p = std::path::PathBuf::from(format!("{}\\{}", userprofile, sub));
            if p.exists() && !dirs.contains(&p) { dirs.push(p); }
        }
    }

    // 4. Backend storage uploads directory in %APPDATA%\Mim\uploads
    if let Ok(appdata) = std::env::var("APPDATA") {
        let p_up = std::path::PathBuf::from(format!("{}\\Mim\\uploads", appdata));
        if p_up.exists() && !dirs.contains(&p_up) { dirs.push(p_up); }
    }

    dirs
}

#[tauri::command]
fn open_downloads_folder(
    app: tauri::AppHandle,
    file_path: Option<String>,
    file_name: Option<String>,
) -> Result<(), String> {
    #[cfg(target_os = "windows")]
    {
        use std::process::Command;

        // 1. If explicit local file_path is passed and exists, select it with raw_arg (preserves /select,"path with spaces")
        if let Some(ref p) = file_path {
            let path = std::path::Path::new(p);
            if path.exists() {
                let mut cmd = Command::new("explorer.exe");
                cmd.raw_arg(format!("/select,\"{}\"", path.to_string_lossy()));
                let _ = cmd.spawn();
                return Ok(());
            }
        }

        // 2. Resolve comprehensive candidate directories from registry (e.g. D:\Users\Lenovo\Downloads)
        let search_dirs = get_windows_user_shell_dirs(&app);

        // 3. Search for file_name in all candidate directories
        if let Some(ref name) = file_name {
            let clean_name = name.trim().trim_matches('"');
            for dir in &search_dirs {
                let candidate = dir.join(clean_name);
                if candidate.exists() {
                    let mut cmd = Command::new("explorer.exe");
                    cmd.raw_arg(format!("/select,\"{}\"", candidate.to_string_lossy()));
                    let _ = cmd.spawn();
                    return Ok(());
                }
            }
        }

        // 4. Fallback: Open the primary directory directly
        let target_dir = search_dirs.first().cloned().unwrap_or_else(|| std::path::PathBuf::from("."));
        let mut cmd = Command::new("explorer.exe");
        cmd.raw_arg(format!("\"{}\"", target_dir.to_string_lossy()));
        let _ = cmd.spawn();
    }
    #[cfg(target_os = "macos")]
    {
        use std::process::Command;
        if let Some(ref p) = file_path {
            let path = std::path::Path::new(p);
            if path.exists() {
                let _ = Command::new("open").arg("-R").arg(path).spawn();
                return Ok(());
            }
        }
        let target_dir = app.path().download_dir().unwrap_or_else(|_| std::path::PathBuf::from("~/Downloads"));
        if let Some(ref name) = file_name {
            let in_dl = target_dir.join(name);
            if in_dl.exists() {
                let _ = Command::new("open").arg("-R").arg(&in_dl).spawn();
                return Ok(());
            }
        }
        let _ = Command::new("open").arg(&target_dir).spawn();
    }
    #[cfg(target_os = "linux")]
    {
        use std::process::Command;
        let target_dir = app.path().download_dir().unwrap_or_else(|_| std::path::PathBuf::from("~/Downloads"));
        let _ = Command::new("xdg-open").arg(&target_dir).spawn();
    }
    Ok(())
}

#[tauri::command]
fn open_file(
    app: tauri::AppHandle,
    file_path: Option<String>,
    file_name: Option<String>,
) -> Result<(), String> {
    #[cfg(target_os = "windows")]
    {
        use std::process::Command;

        // 1. If explicit file_path exists, open with associated Windows default application
        if let Some(ref p) = file_path {
            let path = std::path::Path::new(p);
            if path.exists() {
                let mut cmd = Command::new("cmd.exe");
                cmd.raw_arg(format!("/c start \"\" \"{}\"", path.to_string_lossy()));
                let _ = cmd.spawn();
                return Ok(());
            }
        }

        // 2. Search for file_name across real Windows user shell directories
        let search_dirs = get_windows_user_shell_dirs(&app);

        if let Some(ref name) = file_name {
            let clean_name = name.trim().trim_matches('"');
            for dir in &search_dirs {
                let candidate = dir.join(clean_name);
                if candidate.exists() {
                    let mut cmd = Command::new("cmd.exe");
                    cmd.raw_arg(format!("/c start \"\" \"{}\"", candidate.to_string_lossy()));
                    let _ = cmd.spawn();
                    return Ok(());
                }
            }
        }
    }
    #[cfg(target_os = "macos")]
    {
        use std::process::Command;
        if let Some(ref p) = file_path {
            let path = std::path::Path::new(p);
            if path.exists() {
                let _ = Command::new("open").arg(path).spawn();
                return Ok(());
            }
        }
        if let Some(ref name) = file_name {
            let target_dir = app.path().download_dir().unwrap_or_else(|_| std::path::PathBuf::from("~/Downloads"));
            let in_dl = target_dir.join(name);
            if in_dl.exists() {
                let _ = Command::new("open").arg(&in_dl).spawn();
                return Ok(());
            }
        }
    }
    #[cfg(target_os = "linux")]
    {
        use std::process::Command;
        if let Some(ref p) = file_path {
            let path = std::path::Path::new(p);
            if path.exists() {
                let _ = Command::new("xdg-open").arg(path).spawn();
                return Ok(());
            }
        }
    }
    Ok(())
}

#[tauri::command]
fn download_file(
    app: tauri::AppHandle,
    url: String,
    file_name: String,
) -> Result<String, String> {
    #[cfg(target_os = "windows")]
    {
        use std::process::Command;

        let search_dirs = get_windows_user_shell_dirs(&app);
        let download_dir = search_dirs.iter().find(|p| p.to_string_lossy().to_lowercase().contains("download"))
            .cloned()
            .or_else(|| search_dirs.first().cloned())
            .unwrap_or_else(|| std::path::PathBuf::from("."));

        let clean_name = file_name.trim().trim_matches('"');
        let target_path = download_dir.join(clean_name);

        let full_url = if url.starts_with("http://") || url.starts_with("https://") {
            url
        } else {
            format!("http://127.0.0.1:8080{}", if url.starts_with('/') { "" } else { "/" }) + &url
        };

        // Use Windows built-in curl.exe for instantaneous zero-dependency download
        let mut cmd = Command::new("curl.exe");
        cmd.args(&["-s", "-L", "-o", &target_path.to_string_lossy(), &full_url]);
        let status = cmd.status().map_err(|e| format!("curl execution error: {}", e))?;

        if status.success() && target_path.exists() {
            Ok(target_path.to_string_lossy().to_string())
        } else {
            Err(format!("Download failed with exit code: {:?}", status.code()))
        }
    }
    #[cfg(not(target_os = "windows"))]
    {
        Err("Unsupported OS".to_string())
    }
}

static BACKEND_CHILD: std::sync::Mutex<Option<std::process::Child>> = std::sync::Mutex::new(None);

fn stop_backend() {
    if let Ok(mut lock) = BACKEND_CHILD.lock() {
        if let Some(mut child) = lock.take() {
            let _ = child.kill();
        }
    }
}

fn find_server_binary(app: &tauri::AppHandle) -> Option<std::path::PathBuf> {
    let mut candidates = Vec::new();

    // 1. Current executable directory (production release bundle)
    if let Ok(curr_exe) = std::env::current_exe() {
        if let Some(exe_dir) = curr_exe.parent() {
            candidates.push(exe_dir.join("bin").join("server.exe"));
            candidates.push(exe_dir.join("server.exe"));
            candidates.push(exe_dir.join("resources").join("bin").join("server.exe"));
        }
    }

    // 2. Tauri resource directory
    if let Ok(res_dir) = app.path().resource_dir() {
        candidates.push(res_dir.join("bin").join("server.exe"));
        candidates.push(res_dir.join("server.exe"));
    }

    // 3. Current working directory / dev paths
    candidates.push(std::path::PathBuf::from("bin/server.exe"));
    candidates.push(std::path::PathBuf::from("src-tauri/bin/server.exe"));
    candidates.push(std::path::PathBuf::from("../../backend/server.exe"));
    candidates.push(std::path::PathBuf::from("backend/server.exe"));

    for path in candidates {
        if path.exists() {
            return Some(path);
        }
    }
    None
}

fn start_backend_if_needed(app: &tauri::AppHandle) {
    // If backend is already running on 127.0.0.1:8080, don't start duplicate
    if std::net::TcpStream::connect_timeout(
        &"127.0.0.1:8080".parse().unwrap(),
        std::time::Duration::from_millis(200),
    ).is_ok() {
        return;
    }

    if let Some(server_bin) = find_server_binary(app) {
        // Store user database and uploads in %APPDATA%\Mim or app_data_dir
        let app_data = app.path().app_data_dir().unwrap_or_else(|_| {
            if let Ok(appdata) = std::env::var("APPDATA") {
                std::path::PathBuf::from(format!("{}\\Mim", appdata))
            } else {
                std::path::PathBuf::from(".")
            }
        });
        let _ = std::fs::create_dir_all(&app_data);
        let db_path = app_data.join("mim.db");
        let uploads_dir = app_data.join("uploads");
        let _ = std::fs::create_dir_all(&uploads_dir);

        let mut cmd = std::process::Command::new(&server_bin);
        cmd.arg("-port").arg("8080")
            .arg("-db").arg(db_path.to_string_lossy().to_string())
            .arg("-storage").arg(uploads_dir.to_string_lossy().to_string())
            .current_dir(&app_data);

        let log_file = app_data.join("backend.log");
        if let Ok(file) = std::fs::OpenOptions::new().create(true).write(true).append(true).open(&log_file) {
            if let Ok(err_file) = file.try_clone() {
                cmd.stdout(std::process::Stdio::from(file));
                cmd.stderr(std::process::Stdio::from(err_file));
            }
        }

        #[cfg(target_os = "windows")]
        {
            use std::os::windows::process::CommandExt;
            cmd.creation_flags(0x08000000); // CREATE_NO_WINDOW: prevent CMD popup
        }

        match cmd.spawn() {
            Ok(child) => {
                if let Ok(mut lock) = BACKEND_CHILD.lock() {
                    *lock = Some(child);
                }

                // Wait up to 3 seconds for port 8080 to become ready
                for _ in 0..30 {
                    std::thread::sleep(std::time::Duration::from_millis(100));
                    if std::net::TcpStream::connect_timeout(
                        &"127.0.0.1:8080".parse().unwrap(),
                        std::time::Duration::from_millis(100),
                    ).is_ok() {
                        break;
                    }
                }
            }
            Err(e) => {
                let _ = std::fs::write(app_data.join("backend_error.log"), format!("Failed to spawn backend {:?}: {}", server_bin, e));
            }
        }
    } else {
        let app_data = app.path().app_data_dir().unwrap_or_else(|_| std::path::PathBuf::from("."));
        let _ = std::fs::write(app_data.join("backend_error.log"), "Could not find server.exe binary in any candidate paths");
    }
}

fn main() {
    std::panic::set_hook(Box::new(|info| {
        let msg = format!("PANIC: {}\n", info);
        let _ = std::fs::write("C:\\Users\\Lenovo\\AppData\\Roaming\\Mim\\crash.log", msg);
    }));

    tauri::Builder::default()
        .invoke_handler(tauri::generate_handler![open_downloads_folder, open_file, download_file])
        .setup(|app| {
            // Auto-start backend sidecar if not already running
            start_backend_if_needed(app.handle());

            let show_i = MenuItem::with_id(app, "show", "打开 Mim", true, None::<&str>)?;
            let quit_i = MenuItem::with_id(app, "quit", "退出 Mim", true, None::<&str>)?;
            let menu = Menu::with_items(app, &[&show_i, &quit_i])?;

            let mut tray_builder = TrayIconBuilder::new()
                .menu(&menu)
                .show_menu_on_left_click(false);

            if let Some(icon) = app.default_window_icon() {
                tray_builder = tray_builder.icon(icon.clone());
            }

            let _tray = tray_builder
                .on_menu_event(|app, event| match event.id.as_ref() {
                    "show" => {
                        if let Some(window) = app.get_webview_window("main") {
                            let _ = window.show();
                            let _ = window.unminimize();
                            let _ = window.set_focus();
                        }
                    }
                    "quit" => {
                        stop_backend();
                        app.exit(0);
                    }
                    _ => {}
                })
                .on_tray_icon_event(|tray, event| {
                    if let TrayIconEvent::Click {
                        button: MouseButton::Left,
                        button_state: MouseButtonState::Up,
                        ..
                    } = event
                    {
                        let app = tray.app_handle();
                        if let Some(window) = app.get_webview_window("main") {
                            let _ = window.show();
                            let _ = window.unminimize();
                            let _ = window.set_focus();
                        }
                    }
                })
                .build(app)?;

            if let Some(window) = app.get_webview_window("main") {
                let _ = std::fs::write(
                    "C:\\Users\\Lenovo\\AppData\\Roaming\\Mim\\window_url.log",
                    format!("Window URL: {:?}", window.url()),
                );
            }

            #[cfg(debug_assertions)]
            {
                if let Some(window) = app.get_webview_window("main") {
                    let _ = window.open_devtools();
                }
            }
            Ok(())
        })
        .on_window_event(|window, event| {
            if let tauri::WindowEvent::CloseRequested { api, .. } = event {
                // Minimize to tray on close
                let _ = window.hide();
                api.prevent_close();
            }
        })
        .run(tauri::generate_context!())
        .expect("error while running Mim desktop application");
}
