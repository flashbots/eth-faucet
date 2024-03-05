import { SvelteKitAuth } from "@auth/sveltekit";
import TwitterProvider from '@auth/core/providers/twitter';
import GitHub from '@auth/core/providers/github';

import { env } from "$env/dynamic/private";
import jwt from 'jsonwebtoken';

export const { handle, signIn, signOut } = SvelteKitAuth({
    basePath: "/auth",
    secret: env.AUTH_SECRET,
    trustHost: true,
    useSecureCookies: true,

    providers: [
        GitHub({
            clientId: env.AUTH_GITHUB_ID,
            clientSecret: env.AUTH_GITHUB_SECRET,
        }),
        TwitterProvider({
            clientId: env.AUTH_TWITTER_ID,         // oauth 2.0 client id
            clientSecret: env.AUTH_TWITTER_SECRET, // oauth 2.0 client secret
        }),
    ],

    callbacks: {
        jwt: async({ token, account, profile }) => {
            // account and profile are defined only after fresh authentication
            if (account && profile) {
                token.provider = account.provider;

                switch (account.provider) {
                    case "twitter":
                        token.username = profile.data.username;
                        break;
                    case "github":
                        token.username = profile.login;
                        break;
                }

                token.apiToken = jwt.sign({
                    provider: token.provider,
                    username: token.username,
                }, env.AUTH_SECRET, {
                    audience: env.ORIGIN,
                    expiresIn: "1h",
                });

                return token
            }

            // recreate api token after it's expired
            if (token.apiToken) {
                jwt.verify(token.apiToken, env.AUTH_SECRET, function(err, previousToken) {
                    if (err) {
                        if (err.name == 'TokenExpiredError') {
                            token.apiToken = jwt.sign({
                                provider: token.provider,
                                username: token.username,
                            }, env.AUTH_SECRET, {
                                audience: env.ORIGIN,
                                expiresIn: "1h",
                            });
                        } else {
                            delete token.apiToken;
                        }
                    }
                });

                return token;
            }

            return token;
        },

        session: async ({ session, token }) => {
            if (token) {
                session.user.apiToken = token.apiToken;
                session.user.provider = token.provider;
                session.user.username = token.username;
            }
            return session;
        },
    },
});
