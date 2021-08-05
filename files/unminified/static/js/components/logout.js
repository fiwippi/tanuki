export const name = 'logout';

import * as API from "/static/js/api.js"

export default function () {
    return {
        logout: {
            async ['@click']() {
                await API.Auth.Logout()
                window.location.replace('/login')
            },
        },
    }
}