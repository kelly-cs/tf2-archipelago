import { Model } from '@gen/tf2ap/launcher/v1/form_pb';

/** serverValue is what the launcher holds for one row, wherever it is on the
    model, and undefined for a row the model does not declare. */
export function serverValue(model: Model | undefined, id: string): string | undefined {
  for (const tab of model?.tabs ?? []) {
    for (const field of tab.fields) {
      if (field.id === id) {
        return field.value;
      }
    }
  }
  return undefined;
}
