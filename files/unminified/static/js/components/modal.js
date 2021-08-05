export const name = 'modal';

export default function () {
        return {
            visible: false,

            modal_bg: {
                ['x-show']() {
                    return this.visible
                },
                ['@keyup.escape.document']() {
                    this.hide()
                },
            },
            modal_content: {
                ['x-show']() {
                    return this.visible
                },
                ['@click.away']() {
                    this.hide()
                },
            },
            modal_close: {
                ['@click']() {
                    this.hide()
                },
            },

            show() { this.visible = true },
            hide() { this.visible = false },
        }
}