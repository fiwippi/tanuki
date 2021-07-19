async function authLogin(username, password) {
    let data = { username: username, password: password }
    return await fetch('/auth/login', {
        method: 'POST',
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

async function authLogout() {
    return await fetch('/auth/logout', {
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