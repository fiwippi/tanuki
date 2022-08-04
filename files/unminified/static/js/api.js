export const name = 'api';

const API_URL = "/api/"

Object.prototype.process = function() {
    return new Promise((resolve, reject) => {
        if (this.status === 401 && !this.url.endsWith("/login")) {
            window.location.replace('/')
            reject(this)
        }
        if (this.status >= 400 && this.status <= 500) {
            reject(this)
        }
        resolve(this)
    })
}

Object.prototype.checkJSON = function() {
    return new Promise((resolve, reject) => {
        const contentType = this.headers.get("content-type");
        if (contentType && contentType.indexOf("application/json") !== -1) {
            resolve(this.json())
        }
        resolve(this)
    })
}

async function fetchResource(route, userOptions = {}, form) {
    // Define default options
    const defaultOptions = {
        method: 'GET',
    };
    // Define default headers
    const defaultHeaders = {
        'Content-Type': 'application/json',
    };

    // If we are sending a form, we don't
    // want to use the default content type
    // since it's set automatically
    if (form) {
        delete defaultHeaders["Content-Type"]
    }

    const options = {
        // Merge options
        ...defaultOptions,
        ...userOptions,
        // Merge headers
        headers: {
            ...defaultHeaders,
            ...userOptions.headers,
        },
    };

    return fetch(API_URL + route, options)
        .then(async resp => {
            resp = await resp.process()
            resp = await resp.checkJSON()
            return resp
        })
        .catch(async resp => {
            resp = await resp.checkJSON()
            console.error(resp); throw resp}
        )
}

export class Auth {
    static async Login(username, password) {
        let data = {
            username: username,
            password: password,
        }

        return fetchResource("auth/login", {
            method: 'POST',
            body: JSON.stringify(data),
        })
    }

    static async Logout(username, password) {
        return fetchResource("auth/logout")
    }
}

export class Admin {
    static async ScanLibrary() {
        return fetchResource("admin/library/scan/")
            .then(resp => resp.json())
    }

    static async ViewDB() {
        return fetch(API_URL + "admin/db/view/")
            .then(resp => resp.blob())
            .then(blob => {
                const url = window.URL.createObjectURL(blob);
                const a = document.createElement('a');
                a.style.display = 'none';
                a.href = url;
                a.download = 'db.txt';
                document.body.appendChild(a);
                a.click();
                window.URL.revokeObjectURL(url);
            })
            .catch(error => console.log(error));
    }

    static async GenerateThumbnails() {
        return fetchResource("admin/library/generate-thumbnails/")
            .then(resp => resp.json())
    }

    static async VacuumDB() {
        return fetchResource("admin/db/vacuum/")
            .then(resp => resp.json())
    }

    static async GetMissingItems() {
        return fetchResource("admin/library/missing-items/")
            .then(resp => resp.json())
    }

    static async DeleteMissingItems() {
        return fetchResource("admin/library/missing-items/", {method: 'DELETE'})
    }

    static async Users() {
        return fetchResource("admin/users/")
    }

    static async CreateUser(username, password, type) {
        let data = { username: username, password: password, type: type}
        return fetchResource(`admin/users`, {
            method: 'PUT',
            body: JSON.stringify(data),
        })
    }

    static async DeleteUser(uid) {
        return fetchResource(`admin/user/${uid}/`, {method: 'DELETE'})
    }

    static async EditUser(uid, newUsername, newPassword, newType) {
        let data = { new_username: newUsername, new_password: newPassword, new_type: newType}
        return fetchResource(`admin/user/${uid}/`, {
            method: 'PATCH',
            body: JSON.stringify(data),
        })
    }
}

export class User {
    static async IsAdmin() {
        return fetchResource("user/type/")
            .then(data => {
                return data.type === 'admin';
            })
            .catch(() => {
                return false
            })
    }
}