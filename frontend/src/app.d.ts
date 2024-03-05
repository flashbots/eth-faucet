import "@auth/sveltekit"

declare module "@auth/sveltekit" {
    interface JWT {
        apiToken: string,
        provider: string,
        username: string,
    }

    interface User {
        apiToken: string,
        provider: string,
        username: string,
    }
}
