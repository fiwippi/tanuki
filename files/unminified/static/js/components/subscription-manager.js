export const name = 'subscription-manager';

import * as API from "/static/js/api.js"
import * as Util from "/static/js/util.js"

export default function (subscriptions) {
    return {
        subscriptions: Util.Ensure.Array(subscriptions),

        fmtDate(d) {
            return Util.Fmt.RFCDate(d, true)
        },

        async deleteSub(sid) {
            document.getElementById("subHeading").classList.add("loader")
            await API.Subscription.Delete(sid)
                .then(this.refreshSub())
                .catch(() => {alert("Failed!")})
            document.getElementById("subHeading").classList.remove("loader")
            console.log(this.subscriptions, this.subscriptions.length)
        },

        async refreshSub() {
            await API.Subscription.ViewAll()
                .then(s => {
                    this.subscriptions = Util.Ensure.Array(s)
                })
                .catch(() => {alert("Failed to refresh subscriptions!")})
        },
    }
}