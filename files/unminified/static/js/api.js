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
    }

    static async CheckSubcriptions() {
        return fetchResource("admin/check-subscriptions")
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
    }

    static async VacuumDB() {
        return fetchResource("admin/db/vacuum/")
    }

    static async GetMissingItems() {
        return fetchResource("admin/library/missing-items/")
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

export class Catalog {
    static async SeriesProgress(sid) {
        return fetchResource(`series/${sid}/progress`)
            .then(data => { return data.progress })
    }

    static async EntryProgress(sid, eid) {
        return fetchResource(`series/${sid}/entries/${eid}/progress`)
            .then(data => { return data.progress })
    }

    static async PatchTags(sid, tags) {
        let data = { tags: tags }
        return fetchResource(`series/${sid}/tags`, {
            method: 'PATCH',
            body: JSON.stringify(data),
        })
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
    }
}

export class Mangadex {
    static async Search(title, limit) {
        if (limit == undefined) limit = 10
        return fetchResource(`mangadex/search?title=${title}&limit=${limit}`)
    }

    static async View(uuid) {
        return fetchResource(`mangadex/view/${uuid}`)
    }
}

export class Download {
    static async Chapters(title, chapters, create_subscription) {
        let data = {
            title: title,
            chapters: chapters,
            create_subscription: create_subscription,
        }
        return fetchResource(`manager/download/chapters`, {
            method: 'POST',
            body: JSON.stringify(data),
        })
    }

    static Manager() {
        return new EventSource(`${API_URL}manager/download`)
    }

    static async DeleteAllDownloads() {
        return fetchResource(`manager/download/delete-all-dl`)
    }

    static async DeleteSuccessfulDownloads() {
        return fetchResource(`manager/download/delete-successful-dl`)
    }

    static async RetryFailedDownloads() {
        return fetchResource(`manager/download/retry-failed-dl`)
    }

    static async Pause() {
        return fetchResource(`manager/download/pause-dl`)
    }

    static async Resume() {
        return fetchResource(`manager/download/resume-dl`)
    }

    static async Cancel() {
        return fetchResource(`manager/download/cancel-dl`)
    }
}

export class Subscription {
    static async ViewAll() {
        return fetchResource(`manager/subscription`)
    }

    static async Delete(sid) {
        return fetchResource(`manager/subscription/${sid}`, {method: 'DELETE'})
    }
}