// services/integrations/github.ts
// FIX: Corrected import path for types
import { GitHubRepoListItem } from "../../types";

interface GitHubFileContent {
    name: string;
    path: string;
    type: 'file' | 'dir';
    size?: number;
    content?: string; // base64 encoded
    download_url: string | null;
    url: string;
}

const GITHUB_API_BASE = 'https://api.github.com';

const parseRepoUrl = (url: string): { owner: string; repo: string } | null => {
    try {
        const urlObj = new URL(url);
        if (urlObj.hostname !== 'github.com') return null;
        const pathParts = urlObj.pathname.split('/').filter(Boolean);
        if (pathParts.length < 2) return null;
        return { owner: pathParts[0], repo: pathParts[1].replace('.git', '') };
    } catch (error) {
        return null;
    }
};

export const fetchRepoContents = async (repoUrl: string, pat: string): Promise<string> => {
    const repoInfo = parseRepoUrl(repoUrl);
    if (!repoInfo) {
        throw new Error('Invalid GitHub repository URL.');
    }
    const { owner, repo } = repoInfo;
    // This now delegates to the more comprehensive analysis fetcher
    return fetchRepoForAnalysis(owner, repo, pat);
};


export const listUserRepos = async (username: string, pat: string): Promise<GitHubRepoListItem[]> => {
    if (!username.trim()) {
        throw new Error('GitHub username or organization cannot be empty.');
    }
    const headers = {
        'Authorization': `token ${pat}`,
        'Accept': 'application/vnd.github.v3+json',
    };

    // First, determine if it's a user or an org
    const userTypeRes = await fetch(`${GITHUB_API_BASE}/users/${username}`, { headers });
    if (!userTypeRes.ok) {
        if (userTypeRes.status === 404) throw new Error(`User or organization '${username}' not found.`);
        throw new Error(`Failed to fetch user type: ${userTypeRes.statusText}`);
    }
    const userData = await userTypeRes.json();
    const repoUrl = userData.type === 'Organization' 
        ? `${GITHUB_API_BASE}/orgs/${username}/repos?type=all&sort=updated&per_page=100` 
        : `${GITHUB_API_BASE}/users/${username}/repos?type=all&sort=updated&per_page=100`;

    const reposRes = await fetch(repoUrl, { headers });
    if (!reposRes.ok) {
        throw new Error(`Failed to fetch repositories for '${username}': ${reposRes.statusText}`);
    }

    const reposData: GitHubRepoListItem[] = await reposRes.json();
    return reposData;
};

const CODE_EXTENSIONS = /\.(js|ts|jsx|tsx|py|java|cs|go|rb|php|html|css|scss|md|txt|json|yml|yaml|dockerfile|gitignore|npmrc|env.example)$/i;
const MAX_FILES_TO_FETCH = 50;
const MAX_FILE_SIZE_BYTES = 100 * 1024; // 100KB

async function recursivelyFetchDirectory(
    owner: string,
    repo: string,
    path: string,
    headers: HeadersInit,
    filesFetched: { count: number }
): Promise<string> {
    if (filesFetched.count >= MAX_FILES_TO_FETCH) {
        return '';
    }
    
    let combinedContent = '';

    try {
        const contentsRes = await fetch(`${GITHUB_API_BASE}/repos/${owner}/${repo}/contents/${path}`, { headers });
        if (contentsRes.status === 404) { return ''; } // Directory not found is not an error
        if (contentsRes.status === 401) {
            // Stop further requests if token is invalid
            throw new Error('GitHub token is invalid or has insufficient permissions.');
        }
        if (!contentsRes.ok) {
            console.warn(`Could not fetch contents for path: ${path} - Status: ${contentsRes.status}`);
            return '';
        }

        const contents: GitHubFileContent[] = await contentsRes.json();
        
        for (const item of contents) {
            if (filesFetched.count >= MAX_FILES_TO_FETCH) {
                break;
            }

            if (item.type === 'dir') {
                combinedContent += await recursivelyFetchDirectory(owner, repo, item.path, headers, filesFetched);
            } else if (
                item.type === 'file' && 
                CODE_EXTENSIONS.test(item.name) && 
                item.size && 
                item.size < MAX_FILE_SIZE_BYTES
            ) {
                try {
                    const fileDataRes = await fetch(item.url, { headers });
                    if (!fileDataRes.ok) {
                        console.warn(`Failed to fetch file data for ${item.path}`);
                        continue;
                    }
                    const fileData = await fileDataRes.json();
                    if (fileData.content) {
                        const fileContent = atob(fileData.content);
                        combinedContent += `// / ${item.path} / //\n${fileContent}\n\n---\n\n`;
                        filesFetched.count++;
                    }
                } catch (e) {
                    console.warn(`Error fetching file content for ${item.path}`, e);
                }
            }
        }
    } catch (e) {
        console.error(`Error processing directory ${path}`, e);
        if (e instanceof Error) throw e; // Re-throw to be caught by the caller
    }
    
    return combinedContent;
}

export const fetchRepoForAnalysis = async (owner: string, repo: string, pat: string): Promise<string> => {
    const headers = {
        'Authorization': `token ${pat}`,
        'Accept': 'application/vnd.github.v3+json',
    };
    let combinedContent = '';
    const filesFetched = { count: 0 };
    
    // 1. Fetch README first, as it's often the most important summary
    try {
        const readmeRes = await fetch(`${GITHUB_API_BASE}/repos/${owner}/${repo}/readme`, { headers });
        if (readmeRes.status === 401) throw new Error('GitHub token is invalid or has insufficient permissions.');
        if (readmeRes.ok) {
            const readmeData = await readmeRes.json();
            if (readmeData.content) {
                combinedContent += `// / README.md / //\n${atob(readmeData.content)}\n\n---\n\n`;
                filesFetched.count++;
            }
        }
    } catch (e) {
        console.warn("Could not fetch README.md", e);
        if (e instanceof Error) throw e;
    }
    
    // 2. Recursively fetch from common source and documentation directories
    const sourceDirectories = ['src', 'source', 'lib', 'app', 'docs'];
    for (const dir of sourceDirectories) {
        if (filesFetched.count >= MAX_FILES_TO_FETCH) {
            break;
        }
        combinedContent += await recursivelyFetchDirectory(owner, repo, dir, headers, filesFetched);
    }
    
    if (!combinedContent.trim()) {
        throw new Error('Could not find a README.md or any relevant source files in common directories (src, lib, docs, etc.).');
    }

    return combinedContent.trim();
};