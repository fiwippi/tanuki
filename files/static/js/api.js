// GET /api/user/type
async function apiUserAdmin() {
    return await fetch('/api/user/type', {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
        },
    })
        .then(response => response.json())
        .then(data => {
            return data['type'] === 'admin';
        })
        .catch((error) => {
            console.error(error);
            return false
        })
}

// GET /api/user/progress
async function apiUserProgress(series, entry) {
    let url = '/api/user/progress?'

    if (series.length > 0) {
        url = url + 'series=' +series
    }
    if (entry.length > 0) {
        url = url + '&entry=' + entry
    }

    return await fetch(url, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
        },
    })
        .then(response => response.json())
        .then(data => {
            return data
        })
        .catch((error) => {
            console.error(error);
            return undefined
        })
}

// PATCH /api/user/progress
async function apiPatchUserProgress(series, entry, progress) {
    let url = '/api/user/progress?'

    if (series.length > 0) {
        url = url + 'series=' +series
    }
    if (entry.length > 0) {
        url = url + '&entry=' + entry
    }

    let data = { progress: progress }

    return await fetch(url, {
        method: 'PATCH',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
    })
        .then(response => response.json())
        .then(data => {
            return { success: data['success'] }
        })
        .catch((error) => {
            console.error(error);
            return undefined
        })
}

// GET /api/user/name
async function apiUserName() {
    return await fetch('/api/user/name', {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
        },
    })
        .then(response => response.json())
        .then(data => {
            return data['username']
        })
        .catch((error) => {
            console.error(error);
            return ""
        })
}

// GET /api/admin/users
async function apiAdminUsersView() {
    return await fetch('/api/admin/users', {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
        },
    })
        .then(response => response.json())
        .then(data => {
            return { success: data['success'], users: data['users'] }
        })
        .catch((error) => {
            console.error(error);
            return undefined
        })
}

// PUT /api/admin/users
async function apiAdminUserCreate(username, password, userType) {
    let data = { username: username, password: password, type: userType}
    return await fetch('/api/admin/users', {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
    })
        .then(response => response.json())
        .then(data => {
            return { success: data['success'], message: data['message'] }
        })
        .catch((error) => {
            console.error(error);
            return undefined
        })
}

// GET /api/admin/user/:id
async function apiAdminUserView(usernameHash) {
    return await fetch('/api/admin/user/' + usernameHash, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
        },
    })
        .then(response => response.json())
        .then(data => {
            return { success: data['success'], user: data['user'] }
        })
        .catch((error) => {
            console.error(error);
            return undefined
        })
}

// PATCH /api/admin/user/:id
async function apiAdminUserEdit(usernameHash, newUsername, newPassword, newType) {
    let data = { new_username: newUsername, new_password: newPassword, new_type: newType}
    return await fetch('/api/admin/user/' + usernameHash, {
        method: 'PATCH',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
    })
        .then(response => response.json())
        .then(data => {
            return { success: data['success'], message: data['message'] }
        })
        .catch((error) => {
            console.error(error);
            return undefined
        })
}

// DELETE /api/admin/user/:id
async function apiAdminUserDelete(usernameHash) {
    return await fetch('/api/admin/user/' + usernameHash, {
        method: 'DELETE',
        headers: {
            'Content-Type': 'application/json',
        },
    })
        .then(response => response.json())
        .then(data => {
            return { success: data['success'], message: data['message'] }
        })
        .catch((error) => {
            console.error(error);
            return undefined
        })
}

// GET /api/admin/db
async function apiAdminDB() {
    return await fetch('/api/admin/db', {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
        },
    })
        .then(response => response.json())
        .then(data => {
            return { success: data['success'], db: data['db'] }
        })
        .catch((error) => {
            console.error(error);
            return ""
        })
}

// GET /api/series
async function apiSeriesList() {
    return await fetch('/api/series', {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
        },
    })
        .then(response => response.json())
        .then(data => {
            return { success: data['success'], entries: data['list'] }
        })
        .catch((error) => {
            console.error(error);
            return undefined
        })
}

// GET /api/tag/:tag
async function apiGetSeriesWithTag(tag) {
    return await fetch('/api/tag/' + tag, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
        },
    })
        .then(response => response.json())
        .then(data => {
            return { success: data['success'], entries: data['list'] }
        })
        .catch((error) => {
            console.error(error);
            return undefined
        })
}

// GET /api/series/:sid
async function apiSeries(sid) {
    return await fetch('/api/series/' + sid, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
        },
    })
        .then(response => response.json())
        .then(data => {
            return { success: data['success'], data: data['data'] }
        })
        .catch((error) => {
            console.error(error);
            return undefined
        })
}

// PATCH /api/series/:sid
async function apiPatchSeries(sid, data) {
    return await fetch('/api/series/' + sid, {
        method: 'PATCH',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
    })
        .then(response => response.json())
        .then(data => {
            return data['success']
        })
        .catch((error) => {
            console.error(error);
            return false
        })
}

// GET /api/series/:sid/entries/:eid
async function apiGetEntry(sid, eid) {
    return await fetch('/api/series/' + sid + '/entries/' + eid, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
        },
    })
        .then(response => response.json())
        .then(data => {
            return data
        })
        .catch((error) => {
            console.error(error);
            return undefined
        })
}

// PATCH /api/series/:sid/entries/:eid
async function apiPatchEntry(sid, eid, data) {
    data.chapter = Number(data.chapter)
    data.volume = Number(data.volume)

    return await fetch('/api/series/' + sid + '/entries/' + eid, {
        method: 'PATCH',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
    })
        .then(response => response.json())
        .then(data => {
            return data['success']
        })
        .catch((error) => {
            console.error(error);
            return false
        })
}

// PATCH /api/series/:sid/cover
async function apiPatchSeriesCover(sid, formdata) {
    return await fetch('/api/series/' + sid + '/cover', {
        // ContentType is NOT set when sending file in form
        method: 'PATCH',
        body: formdata,
    })
        .then(response => response.json())
        .then(data => {
            return data['success']
        })
        .catch((error) => {
            console.error(error);
            return false
        })
}

// PATCH /api/series/:sid/entries/:eid/cover
async function apiPatchEntryCover(sid, eid, formdata) {
    return await fetch('/api/series/' + sid + '/entries/' + eid + '/cover', {
        // ContentType is NOT set when sending file in form
        method: 'PATCH',
        body: formdata,
    })
        .then(response => response.json())
        .then(data => {
            return data['success']
        })
        .catch((error) => {
            console.error(error);
            return false
        })
}

// DELETE /api/series/:sid/cover
async function apiDeleteSeriesCover(sid) {
    return await fetch('/api/series/' + sid + '/cover', {
        method: 'DELETE',
        headers: {
            'Content-Type': 'application/json',
        },
    })
        .then(response => response.json())
        .then(data => {
            return data['success']
        })
        .catch((error) => {
            console.error(error);
            return false
        })
}

// DELETE /api/series/:sid/cover
async function apiDeleteEntryCover(sid, eid) {
    return await fetch('/api/series/' + sid + '/entries/' + eid + '/cover', {
        method: 'DELETE',
        headers: {
            'Content-Type': 'application/json',
        },
    })
        .then(response => response.json())
        .then(data => {
            return data['success']
        })
        .catch((error) => {
            console.error(error);
            return false
        })
}

// GET /api/series/:sid/entries
async function apiSeriesEntries(sid) {
    let url = '/api/series/' + sid + '/entries'

    return await fetch(url, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
        },
    })
        .then(response => response.json())
        .then(data => {
            return { success: data['success'], entries: data['list'], series_hash: data["series_hash"]}
        })
        .catch((error) => {
            console.error(error);
            return undefined
        })
}

// GET /api/admin/library/scan
async function apiAdminLibraryScan() {
    return await fetch('/api/admin/library/scan', {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
        },
    })
        .then(response => response.json())
        .then(data => {
            return { success: data['success'], message: data['message'] }
        })
        .catch((error) => {
            console.error(error);
            return undefined
        })
}

// GET /api/admin/library/generate-thumbnails
async function apiAdminLibraryGenerateThumbnails() {
    return await fetch('/api/admin/library/generate-thumbnails', {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
        },
    })
        .then(response => response.json())
        .then(data => {
            return { success: data['success'], message: data['message'] }
        })
        .catch((error) => {
            console.error(error);
            return undefined
        })
}

// GET /api/admin/library/missing-entries
async function apiAdminLibraryMissingEntries() {
    return await fetch('/api/admin/library/missing-entries', {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
        },
    })
        .then(response => response.json())
        .then(data => {
            return { success: data['success'], entries: data['entries'] }
        })
        .catch((error) => {
            console.error(error);
            return undefined
        })
}

// DELETE /api/admin/library/missing-entries
async function apiDeleteAdminLibraryMissingEntries() {
    return await fetch('/api/admin/library/missing-entries', {
        method: 'DELETE',
        headers: {
            'Content-Type': 'application/json',
        },
    })
        .then(response => response.json())
        .then(data => {
            return { success: data['success'] }
        })
        .catch((error) => {
            console.error(error);
            return false
        })
}

// PATCH /api/series/:sid/tags
async function apiPatchSeriesTags(sid, tags) {
    if (sid === undefined) {
        return undefined
    } else if (sid.length === 0) {
        return undefined
    }

    let data = { tags: tags }
    return await fetch('/api/series/' + sid + '/tags', {
        method: 'PATCH',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
    })
        .then(response => response.json())
        .then(data => {
            return { success: data['success'] }
        })
        .catch((error) => {
            console.error(error);
            return false
        })
}

// GET /api/tags
async function apiTags() {
    return await fetch('/api/tags', {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
        },
    })
        .then(response => response.json())
        .then(data => {
            return { success: data['success'], tags: data['tags'] }
        })
        .catch((error) => {
            console.error(error);
            return undefined
        })
}