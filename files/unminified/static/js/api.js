export const name = 'api';

const API_URL = "/api/"

Object.prototype.ensureAuthed = function() {
    return new Promise((resolve, reject) => {
        if (this.status === 401 && !this.url.endsWith("/login")) {
            window.location.replace('/')
            reject()
        }
        resolve(this)
    })
}

Object.prototype.ensureSuccess = function() {
    return new Promise((resolve, reject) => {
        if (!this.success) {
            if (this.message && this.message.length > 0) {
                console.error(this.message)
            } else {
                console.error("failed:", this)
            }
            reject()
            return
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
        .then(response => response.ensureAuthed())
        .then(response => response.json())
        .catch(error => {throw error})
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
            .then(resp => resp.ensureSuccess())
    }

    static async Logout(username, password) {
        return fetchResource("auth/logout")
    }
}

export class User {
    static async IsAdmin() {
        return fetchResource("user/type/")
            .then(resp => resp.ensureSuccess())
            .then(data => {
                return data.type === 'admin';
            })
            .catch(() => {
                return false
            })
    }

    static async Name() {
        return fetchResource("user/name/")
            .then(resp => resp.ensureSuccess())
            .then(data => {
                return data.name
            })
            .catch(() => {
                return ""
            })
    }
}

export class Admin {
    static async ScanLibrary() {
        return fetchResource("admin/library/scan/")
            .then(resp => resp.ensureSuccess())
    }

    static async ViewDB() {
        return fetch(API_URL + "admin/db")
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
            .then(resp => resp.ensureSuccess())
    }

    static async GetMissingItems() {
        return fetchResource("admin/library/missing-items/")
            .then(resp => resp.ensureSuccess())
    }

    static async DeleteMissingItems() {
        return fetchResource("admin/library/missing-items/", {method: 'DELETE'})
            .then(resp => resp.ensureSuccess())
    }

    static async Users() {
        return fetchResource("admin/users/")
            .then(resp => resp.ensureSuccess())
    }

    static async CreateUser(username, password, type) {
        let data = { username: username, password: password, type: type}
        return fetchResource(`admin/users`, {
            method: 'PUT',
            body: JSON.stringify(data),
        })
            .then(resp => resp.ensureSuccess())
    }

    static async DeleteUser(uid) {
        return fetchResource(`admin/user/${uid}/`, {method: 'DELETE'})
            .then(resp => resp.ensureSuccess())
    }

    static async EditUser(uid, newUsername, newPassword, newType) {
        let data = { new_username: newUsername, new_password: newPassword, new_type: newType}
        return fetchResource(`admin/user/${uid}/`, {
            method: 'PATCH',
            body: JSON.stringify(data),
        })
            .then(resp => resp.ensureSuccess())
    }

}

export class Catalog {
    static async Series(sid) {
        return fetchResource(`series/${sid}`)
            .then(resp => resp.ensureSuccess())
            .then(data => { return data.data })
    }

    static async Entries(sid) {
        return fetchResource(`series/${sid}/entries`)
            .then(resp => resp.ensureSuccess())
            .then(data => {
                return { entries: data.list, series_hash: data.series_hash }
            })
    }

    static async SeriesProgress(sid) {
        return fetchResource(`series/${sid}/progress`)
            .then(resp => resp.ensureSuccess())
            .then(data => { return data.progress })
    }

    static async EntryProgress(sid, eid) {
        return fetchResource(`series/${sid}/entries/${eid}/progress`)
            .then(resp => resp.ensureSuccess())
            .then(data => { return data.progress })
    }

    static async PatchSeries(sid, title, author, date_released) {
        let data = {
            title: title,
            author: author,
            date_released: date_released
        }

        return fetchResource(`series/${sid}`, {
            method: 'PATCH',
            body: JSON.stringify(data),
        })
            .then(resp => resp.ensureSuccess())
    }

    static async PatchTags(sid, tags) {
        let data = { tags: tags }
        return fetchResource(`series/${sid}/tags`, {
            method: 'PATCH',
            body: JSON.stringify(data),
        })
            .then(resp => resp.ensureSuccess())
    }

    static async PatchProgress(sid, eid, progress) {
        let url
        if (sid.length > 0) {
            url = `series/${sid}/progress`
        }
        if (eid.length > 0) {
            url = `series/${sid}/entries/${eid}/progress`
        }

        let data = { progress: progress }

        return fetchResource(url, {
            method: 'PATCH',
            body: JSON.stringify(data),
        })
            .then(resp => resp.ensureSuccess())
    }

    static async PatchSeriesCover(sid, file, filename) {
        let url = `series/${sid}/cover`
        return Catalog.patchCover(url, file, filename)
    }

    static async DeleteSeriesCover(sid) {
        return fetchResource(`series/${sid}/cover`, {method: 'DELETE'})
            .then(resp => resp.ensureSuccess())
    }

    static async PatchEntry(sid, eid, title, author, date_released, chapter, volume) {
        let data = {
            title: title,
            author: author,
            date_released: date_released,
            chapter: Number(chapter),
            volume: Number(volume),
        }

        return fetchResource(`series/${sid}/entries/${eid}`, {
            method: 'PATCH',
            body: JSON.stringify(data),
        })
            .then(resp => resp.ensureSuccess())
    }

    static async PatchEntryCover(sid, eid, file, filename) {
        let url = `series/${sid}/entries/${eid}/cover`
        return Catalog.patchCover(url, file, filename)
    }

    static async DeleteEntryCover(sid, eid) {
        return fetchResource(`series/${sid}/entries/${eid}/cover`, {method: 'DELETE'})
            .then(resp => resp.ensureSuccess())
    }

    static async patchCover(url, file, filename) {
        const form = new FormData();
        form.append('file', file);
        form.append('filename', filename);

        return fetchResource(url, {
            method: 'PATCH',
            body: form,
        }, true)
            .then(resp => resp.ensureSuccess())
    }
}

export class Download {
    static async Chapters(chapters) {
        return fetchResource(`download/chapters`, {
            method: 'POST',
            body: JSON.stringify(chapters),
        })
            .then(resp => resp.ensureSuccess())
    }

    static Manager() {
        return new EventSource(`${API_URL}download/manager`)
    }

    static async DeleteFinishedTasks() {
        return fetchResource(`download/manager/delete-finished-tasks`)
            .then(resp => resp.ensureSuccess())
    }

    static async Pause() {
        return fetchResource(`download/manager/pause-downloads`)
            .then(resp => resp.ensureSuccess())
    }

    static async Resume() {
        return fetchResource(`download/manager/resume-downloads`)
            .then(resp => resp.ensureSuccess())
    }

    static async Cancel() {
        return fetchResource(`download/manager/cancel-downloads`)
            .then(resp => resp.ensureSuccess())
    }
}