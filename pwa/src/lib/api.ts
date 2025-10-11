/**
 * REST API client for Agent Payment endpoints
 */

// Use proxy in dev, direct API in production
const BASE_URL = import.meta.env.DEV ? '/api' : 'https://api.agentpmt.com';

export interface ToolParameter {
  name: string;
  description?: string;
  type: string;
  required: boolean;
}

export interface ToolFunction {
  name: string;
  description: string;
  parameters: {
    type: 'object';
    properties: Record<string, any>;
    required?: string[];
  };
}

export interface ToolRecord {
  type: 'function';
  function: ToolFunction;
  'x-prepaid-balance'?: number;
  'x-pricing'?: {
    cost?: number;
    currency?: string;
  };
}

export interface FetchToolsResponse {
  success: boolean;
  preprompt?: string;
  example?: unknown;
  tools: ToolRecord[];
  pagination?: {
    page: number;
    page_size: number;
    total: number;
  };
}

export interface PurchaseToolResponse {
  success: boolean;
  result: unknown;
  cost?: number;
  balance?: number;
}

export interface ApiCredentials {
  apiKey: string;
  budgetKey: string;
  auth?: string;
}

/**
 * Fetch available tools from the API
 */
export async function fetchTools(
  credentials: ApiCredentials,
  page = 1,
  pageSize = 50
): Promise<FetchToolsResponse> {
  const url = `${BASE_URL}/products/fetch?page=${page}&page_size=${pageSize}`;

  const response = await fetch(url, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
      'X-API-Key': credentials.apiKey,
      'X-Budget-Key': credentials.budgetKey,
      ...(credentials.auth ? { 'Authorization': credentials.auth } : {})
    }
  });

  if (!response.ok) {
    const contentType = response.headers.get('content-type');
    let errorMessage = `Failed to fetch tools: ${response.status} ${response.statusText}`;

    // Try to get error details from response
    try {
      if (contentType?.includes('application/json')) {
        const errorData = await response.json();
        errorMessage += `\n${JSON.stringify(errorData, null, 2)}`;
      } else {
        const textError = await response.text();
        errorMessage += `\n${textError.substring(0, 200)}`;
      }
    } catch (e) {
      // Ignore parse errors, use basic message
    }

    throw new Error(errorMessage);
  }

  // Verify response is JSON before parsing
  const contentType = response.headers.get('content-type');
  if (!contentType?.includes('application/json')) {
    const text = await response.text();
    throw new Error(`Expected JSON response, got: ${contentType}\n${text.substring(0, 200)}`);
  }

  return response.json();
}

/**
 * Execute a tool via the API
 */
export async function purchaseTool(
  credentials: ApiCredentials,
  productId: string,
  parameters: Record<string, unknown>
): Promise<PurchaseToolResponse> {
  const url = `${BASE_URL}/products/purchase`;

  const response = await fetch(url, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-API-Key': credentials.apiKey,
      'X-Budget-Key': credentials.budgetKey
    },
    body: JSON.stringify({
      product_id: productId,
      parameters
    })
  });

  if (!response.ok) {
    const contentType = response.headers.get('content-type');
    let errorMessage = `Failed to purchase tool: ${response.status} ${response.statusText}`;

    try {
      if (contentType?.includes('application/json')) {
        const errorData = await response.json();
        errorMessage += `\n${JSON.stringify(errorData, null, 2)}`;
      } else {
        const textError = await response.text();
        errorMessage += `\n${textError.substring(0, 200)}`;
      }
    } catch (e) {
      // Ignore parse errors
    }

    throw new Error(errorMessage);
  }

  const contentType = response.headers.get('content-type');
  if (!contentType?.includes('application/json')) {
    const text = await response.text();
    throw new Error(`Expected JSON response, got: ${contentType}\n${text.substring(0, 200)}`);
  }

  return response.json();
}
