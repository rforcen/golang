/**
 * Exception thrown when quickHull.QuickHull3D encounters an internal error.
 */
class InternalErrorException extends RuntimeException {
    public InternalErrorException(String msg) {
        super(msg);
    }
}
