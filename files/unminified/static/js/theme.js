// Theme validator

const validThemeSetting = (theme) => {
    return ['light', 'dark'].indexOf(theme) >= 0;
};

// Saving and loading from storage

const loadTheme = () => {
    let str = localStorage.getItem('theme');
    if (!str || !validThemeSetting(str)) str = 'light';
    return str;
};

const saveTheme = (theme) => {
    if (!validThemeSetting(theme)) {
        console.error("tried to save invalid theme:", theme)
        theme = 'light';
    }
    localStorage.setItem("theme", theme)
};

// Public UI functions

const toggleTheme = () => {
    const theme = loadTheme();
    const newTheme = theme === 'dark' ? 'light' : 'dark';
    saveTheme(newTheme)
    setTheme(newTheme);
};

const setTheme = (theme) => {
    if (!theme) theme = loadTheme();
    if (theme === 'dark') {
        document.documentElement.classList.add('dark')
    } else {
        document.documentElement.classList.remove('dark')
    }
};

// Set the theme on page load
setTheme()