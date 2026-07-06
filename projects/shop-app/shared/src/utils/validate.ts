import { type ZodSchema } from "zod";

/**
 * Validate data against a Zod schema.
 * Returns the parsed data on success, or throws a formatted error on failure.
 */
export function validate<T>(schema: ZodSchema<T>, data: unknown): T {
  const result = schema.safeParse(data);

  if (!result.success) {
    const errors = result.error.issues.map(
      (issue) => `${issue.path.join(".")}: ${issue.message}`,
    );
    throw new Error(`Validation failed:\n${errors.join("\n")}`);
  }

  return result.data;
}

/**
 * Validate data against a Zod schema — returns a result tuple instead of
 * throwing, making it convenient to use in components and hooks.
 *
 * @returns [data, null] on success, [null, Error] on failure
 */
export function safeValidate<T>(
  schema: ZodSchema<T>,
  data: unknown,
): [T, null] | [null, Error] {
  const result = schema.safeParse(data);

  if (!result.success) {
    const errors = result.error.issues.map(
      (issue) => `${issue.path.join(".")}: ${issue.message}`,
    );
    return [null, new Error(`Validation failed:\n${errors.join("\n")}`)];
  }

  return [result.data, null];
}
