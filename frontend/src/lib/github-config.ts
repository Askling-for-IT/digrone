export const GITHUB_CONFIG = {
    owner: "Askling-for-IT",
    repo: "digrone",
    branch: "main",
    dataPath: "frontend/public/data"
};

export const getGitHubToken = (): string | null => {
    return localStorage.getItem('github_token');
};

export const setGitHubToken = (token: string): void => {
    localStorage.setItem('github_token', token);
};
