export const name = 'navbar';

import * as Util from "/static/js/util.js"

export default function () {
    return {
        showNavbar: true,
        smallMedia: false,

        init() {
            if (window.innerWidth <= Util.SmallMediaLimit) {
                this.showNavbar = false
                this.smallMedia = true
            }
        }
    }
}